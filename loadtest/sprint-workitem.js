// k6 load test: sprint + work item flow against the local dev stack.
//
//   make loadtest         full ramping run
//   make loadtest-smoke   1 VU, 1 iteration (checks the flow works)
//
// Each VU logs in once (re-logging in before its token expires), then loops:
// create sprint -> create work item (WS) -> put item in sprint (WS) ->
// get sprint -> get sprint metrics.
// All VUs share one project created in setup(). teardown() deletes the
// sprints and work items; the project itself must be deleted by hand (#22).

import http from 'k6/http';
import ws from 'k6/ws';
import { check, fail } from 'k6';
import { Trend } from 'k6/metrics';
import exec from 'k6/execution';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8083';
const WS_URL = __ENV.WS_URL || BASE_URL.replace(/^http/, 'ws') + '/ws/work-items';
const KC_URL = __ENV.KC_URL || 'http://localhost:8080';
const KC_REALM = __ENV.KC_REALM || 'chura';
const KC_CLIENT_ID = __ENV.KC_CLIENT_ID || 'chura-auth-client';
const PASSWORD = __ENV.TEST_PASSWORD || 'test1234';
// First user creates the project; the rest are added as members.
const USERS = (__ENV.TEST_USERS || 'tonnam,pim,viva').split(',');

const TOKEN_REFRESH_MARGIN_MS = 30 * 1000;
const WS_TIMEOUT_MS = 10 * 1000;

const wsOpDuration = new Trend('ws_op_duration', true);

export const options = {
  scenarios: {
    sprint_flow: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 10 },
        { duration: '2m', target: 10 },
        { duration: '1m', target: 30 },
        { duration: '2m', target: 30 },
        { duration: '30s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    checks: ['rate>0.99'],
    'http_req_duration{name:create_sprint}': ['p(95)<500'],
    'http_req_duration{name:get_sprint}': ['p(95)<500'],
    'http_req_duration{name:get_sprint_metrics}': ['p(95)<500'],
    'ws_op_duration{op:create}': ['p(95)<500'],
    'ws_op_duration{op:update}': ['p(95)<500'],
  },
};

// ---- auth ------------------------------------------------------------------

function login(username) {
  const res = http.post(
    `${KC_URL}/realms/${KC_REALM}/protocol/openid-connect/token`,
    { grant_type: 'password', client_id: KC_CLIENT_ID, username, password: PASSWORD },
    { tags: { name: 'keycloak_login' } },
  );
  if (!check(res, { 'login 200': (r) => r.status === 200 })) {
    fail(`login failed for ${username}: ${res.status} ${res.body}`);
  }
  const body = res.json();
  return { token: body.access_token, expiresAt: Date.now() + body.expires_in * 1000 };
}

// Per-VU session: module-level state is separate for every VU.
let session = null;

function tokenFor(username) {
  if (!session || Date.now() > session.expiresAt - TOKEN_REFRESH_MARGIN_MS) {
    session = login(username);
  }
  return session.token;
}

function authParams(token, name) {
  return {
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    tags: { name },
  };
}

// ---- setup / teardown --------------------------------------------------------

function apiOrFail(res, expectedStatus, what) {
  if (res.status !== expectedStatus) {
    fail(`${what}: expected ${expectedStatus}, got ${res.status} ${res.body}`);
  }
  return res.json();
}

export function setup() {
  const owner = login(USERS[0]).token;

  const templates = apiOrFail(
    http.get(`${BASE_URL}/api/v1/templates`, authParams(owner, 'setup')), 200, 'list templates');
  if (!templates.length) fail('no templates available');

  const project = apiOrFail(
    http.post(`${BASE_URL}/api/v1/projects`, JSON.stringify({
      name: `k6-loadtest-${Date.now()}`,
      description: 'Created by k6 load test. Safe to delete.',
      template_id: templates[0].id,
    }), authParams(owner, 'setup')),
    201, 'create project');

  for (const username of USERS.slice(1)) {
    const me = apiOrFail(
      http.get(`${BASE_URL}/api/v1/whoami`, authParams(login(username).token, 'setup')), 200, 'whoami');
    apiOrFail(
      http.post(`${BASE_URL}/api/v1/projects/${project.id}/members`,
        JSON.stringify({ email: me.email }), authParams(owner, 'setup')),
      201, `add member ${username}`);
  }

  console.log(`project ${project.id} ready for ${USERS.length} users`);
  return { projectId: project.id };
}

export function teardown({ projectId }) {
  const owner = login(USERS[0]).token;

  const sprints = apiOrFail(
    http.get(`${BASE_URL}/api/v1/sprints?project_id=${projectId}`, authParams(owner, 'teardown')),
    200, 'list sprints');
  for (const sprint of sprints) {
    http.del(`${BASE_URL}/api/v1/sprints/${sprint.id}`, null, authParams(owner, 'teardown'));
  }

  const deleted = deleteAllWorkItems(projectId);
  console.log(`cleaned ${sprints.length} sprints, ${deleted} work items; ` +
    `delete project ${projectId} (and its Keycloak group) by hand`);
}

function deleteAllWorkItems(projectId) {
  let pending = [];
  let deleted = 0;
  ws.connect(WS_URL, { tags: { name: 'teardown' } }, (socket) => {
    const deleteNext = () => {
      if (!pending.length) return socket.close();
      socket.send(JSON.stringify({ operation: 'delete', id: pending.shift() }));
    };
    socket.on('open', () => socket.send(JSON.stringify({ operation: 'list', project_id: projectId })));
    socket.on('message', (raw) => {
      const msg = JSON.parse(raw);
      if (msg.operation === 'listed') pending = (msg.work_items || []).map((item) => item.id);
      if (msg.operation === 'deleted' && !msg.error) deleted++;
      deleteNext();
    });
    socket.setTimeout(() => socket.close(), 5 * 60 * 1000);
  });
  return deleted;
}

// ---- VU iteration ------------------------------------------------------------

export default function sprintFlow({ projectId }) {
  const username = USERS[(exec.vu.idInTest - 1) % USERS.length];
  const label = `k6-${exec.vu.idInTest}-${exec.vu.iterationInScenario}`;

  const sprintRes = http.post(`${BASE_URL}/api/v1/sprints`,
    JSON.stringify({ project_id: projectId, name: label, team: 'k6' }),
    authParams(tokenFor(username), 'create_sprint'));
  if (!check(sprintRes, { 'create sprint 201': (r) => r.status === 201 })) return;
  const sprintId = sprintRes.json().id;

  if (!createWorkItemInSprint(projectId, sprintId, label)) return;

  check(http.get(`${BASE_URL}/api/v1/sprints/${sprintId}`,
    authParams(tokenFor(username), 'get_sprint')),
  { 'get sprint 200': (r) => r.status === 200 });

  check(http.get(`${BASE_URL}/api/v1/sprints/${sprintId}/metrics`,
    authParams(tokenFor(username), 'get_sprint_metrics')),
  { 'get metrics 200': (r) => r.status === 200 });
}

// Creates a work item, then updates it with sprint_id, over one WS connection.
function createWorkItemInSprint(projectId, sprintId, title) {
  let ok = false;
  let sentAt = 0;
  const send = (socket, message) => {
    sentAt = Date.now();
    socket.send(JSON.stringify(message));
  };

  const res = ws.connect(WS_URL, { tags: { name: 'ws_work_items' } }, (socket) => {
    socket.on('open', () => send(socket, {
      operation: 'create',
      work_item: { project_id: projectId, title, type: 'task', status: 'to_do', priority: 'medium' },
    }));
    socket.on('message', (raw) => {
      const msg = JSON.parse(raw);
      wsOpDuration.add(Date.now() - sentAt, { op: msg.operation === 'created' ? 'create' : 'update' });
      if (!check(msg, { [`ws ${msg.operation} ok`]: (m) => !m.error })) return socket.close();

      if (msg.operation === 'created') {
        send(socket, { operation: 'update', id: msg.work_item.id, work_item: { ...msg.work_item, sprint_id: sprintId } });
      } else {
        ok = check(msg, { 'work item in sprint': (m) => m.work_item.sprint_id === sprintId });
        socket.close();
      }
    });
    socket.setTimeout(() => socket.close(), WS_TIMEOUT_MS);
  });

  check(res, { 'ws upgrade 101': (r) => r?.status === 101 });
  return ok;
}
