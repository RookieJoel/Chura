/* Flowline v2 - parallax stage. No dependencies. */
(function () {
  'use strict';

  var GRID = 56;                                   // keep in sync with flowline.css
  var reduce = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  // layer, class, left %, top %, size px, rotation deg, colour
  var SHAPES = [
    ['b', 'fl-ring',          14, 16, 380,   0, 'pink'],
    ['b', 'fl-ring',          36, 64, 240,   0, 'accent'],
    ['b', 'fl-ring',          58, 10, 120,   0, 'pink'],
    ['b', 'fl-slab',          40, 30, 320, -35, 'accent'],
    ['b', 'fl-slab fl-fill',   8, 56, 380, -35, 'pink'],
    ['c', 'fl-cross',         24, 44,  22,   0, 'pink'],
    ['c', 'fl-cross',         50, 78,  18,  45, 'accent'],
    ['c', 'fl-dot',           12, 34,  10,   0, 'pink'],
    ['c', 'fl-dot',           46, 20,   8,   0, 'accent'],
    ['c', 'fl-dot',           28, 86,  12,   0, 'pink'],
    ['c', 'fl-bar',           60, 52,  90, -35, 'pink'],
    ['c', 'fl-bar',           20,  8,  60, -35, 'accent'],
    ['c', 'fl-bokeh',          4, 74, 260,   0, 'pink'],
    ['c', 'fl-bokeh',         52, 38, 200,   0, 'accent']
  ];

  function rand(a, b) { return a + Math.random() * (b - a); }
  function div(cls) { var d = document.createElement('div'); d.className = cls; return d; }
  function layer(depth) { var l = div('fl-layer'); l.depth = depth; return l; }

  function build() {
    if (document.querySelector('.fl-stage')) return;

    var brand = document.getElementById('kc-header-wrapper');
    var name = brand ? brand.textContent.trim() : '';

    var stage = div('fl-stage');
    stage.setAttribute('aria-hidden', 'true');

    var wash = layer(6),  plane = layer(14), hero = layer(22), far = layer(34), near = layer(62);
    wash.appendChild(div('fl-wash'));

    var field = div('fl-field');
    plane.appendChild(field);

    if (name) {
      var h = div('fl-hero'); var b = document.createElement('b'); b.textContent = name;
      h.appendChild(b); hero.appendChild(h);
    }

    SHAPES.forEach(function (s) {
      var e = div('fl-s ' + s[1] + (s[6] === 'accent' ? ' is-accent' : ''));
      e.style.left = s[2] + '%';
      e.style.top = s[3] + '%';
      e.style.setProperty('--s', s[4] + 'px');
      e.style.setProperty('--r', s[5] + 'deg');
      e.style.setProperty('--bd', rand(7, 12).toFixed(1) + 's');
      e.style.setProperty('--bdelay', (-rand(0, 8)).toFixed(1) + 's');
      (s[0] === 'b' ? far : near).appendChild(e);
    });

    [wash, plane, hero, far, near].forEach(function (l) { stage.appendChild(l); });
    document.body.insertBefore(stage, document.body.firstChild);

    if (reduce) return;                            // static composition only

    // drifting lines, snapped to grid rows
    var rows = Math.max(1, Math.floor(field.offsetHeight / GRID));
    var count = window.innerWidth < 640 ? 14 : 26;
    for (var i = 0; i < count; i++) {
      var ln = document.createElement('i');
      ln.className = 'kc-line' + (i % 4 === 0 ? ' kc-line--accent' : '');
      ln.style.top = (Math.floor(Math.random() * rows) * GRID) + 'px';
      ln.style.left = rand(5, 80).toFixed(1) + '%';
      ln.style.setProperty('--len', rand(140, 360).toFixed(0) + 'px');
      ln.style.setProperty('--travel', rand(500, 1100).toFixed(0) + 'px');
      ln.style.setProperty('--dur', rand(7, 15).toFixed(1) + 's');
      ln.style.setProperty('--delay', (-rand(0, 15)).toFixed(1) + 's');
      ln.style.setProperty('--peak', rand(.45, .95).toFixed(2));
      field.appendChild(ln);
    }

    // parallax: pointer drives it; when idle (or on touch) it drifts on its own
    var layers = [wash, plane, hero, far, near];
    var tx = 0, ty = 0, cx = 0, cy = 0, lastMove = -1e9;

    window.addEventListener('pointermove', function (e) {
      tx = (e.clientX / window.innerWidth - .5) * 2;
      ty = (e.clientY / window.innerHeight - .5) * 2;
      lastMove = performance.now();
    }, { passive: true });

    (function frame(now) {
      if (now - lastMove > 2500) {
        tx = Math.sin(now / 5200) * .55;
        ty = Math.cos(now / 6800) * .40;
      }
      cx += (tx - cx) * .06;
      cy += (ty - cy) * .06;
      for (var k = 0; k < layers.length; k++) {
        layers[k].style.transform =
          'translate3d(' + (-cx * layers[k].depth).toFixed(2) + 'px,' + (-cy * layers[k].depth).toFixed(2) + 'px,0)';
      }
      window.requestAnimationFrame(frame);
    })(performance.now());
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', build);
  else build();
})();
