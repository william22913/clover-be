(function () {
  'use strict';

  /* ---------------------------------------------------------------- data */

  var IMG = 'assets/img/';
  var CLIP = 'assets/video/clover-sample.mp4';

  var FILMS = [
    {
      id: 'halcyon', title: 'Halcyon Bay', year: 2023, runtime: '2h 06m',
      still: 'still-halcyon', poster: 'p-halcyon',
      short: 'A marine biologist comes back to the town that took her brother, and finds the tide still keeping his secrets.',
      rail: 'A marine biologist comes back to the town that took her brother.'
    },
    {
      id: 'quiet', title: 'The Quiet Province', year: 2025, runtime: '1h 44m',
      still: 'still-quiet', poster: 'p-quiet',
      short: 'A schoolteacher keeps a village’s last winter alive by writing letters no one will ever post.',
      rail: 'A village’s last winter, kept alive in letters nobody posts.'
    },
    {
      id: 'vesper', title: 'Vesper Light', year: 2022, runtime: '1h 51m',
      still: 'still-vesper', poster: 'p-vesper',
      short: 'Two strangers miss the same last ferry and spend one night deciding whether to miss it on purpose.'
    },
    {
      id: 'cartographer', title: 'The Cartographer', year: 2024, runtime: '1h 58m',
      still: 'still-quiet', poster: 'p-cartographer',
      short: 'One woman sets out to map the last unmapped valley in Europe, one ridge at a time.'
    },
    {
      id: 'meridian', title: 'Meridian Line', year: 2021, runtime: '2h 12m',
      still: 'still-vesper', poster: 'p-meridian',
      short: 'A night train, a stolen passport, and eight hours to the border.'
    },
    {
      id: 'glasshouse', title: 'Glasshouse Summer', year: 2024, runtime: '1h 39m',
      still: 'still-halcyon', poster: 'p-glasshouse',
      short: 'Three siblings, one greenhouse, and the summer their mother’s will is read.'
    },
    {
      id: 'fathoms', title: 'Nine Fathoms', year: 2023, runtime: '1h 47m',
      still: 'still-halcyon', poster: 'p-fathoms',
      short: 'A deep-sea station loses contact with the surface for eleven minutes.'
    },
    {
      id: 'lowcountry', title: 'Low Country', year: 2022, runtime: '2h 04m',
      still: 'still-quiet', poster: 'p-lowcountry',
      short: 'A family of shrimpers fights the water, the bank, and each other for one season.'
    },
    {
      id: 'saltyear', title: 'The Salt Year', year: 2025, runtime: '1h 36m',
      still: 'still-vesper', poster: 'p-saltyear',
      short: 'A year on the road with a band that never made it and never stopped.'
    },
    {
      id: 'lanterns', title: 'Paper Lanterns', year: 2024, runtime: '1h 42m',
      still: 'still-vesper', poster: 'p-lanterns',
      short: 'A projectionist falls for the woman who returns to the same empty seat.'
    },
    {
      id: 'harbour', title: 'Harbour Lights', year: 2021, runtime: '1h 55m',
      still: 'still-halcyon', poster: 'p-harbour',
      short: 'The last show at a seaside theatre, on the night the lights went out.'
    },
    {
      id: 'understudy', title: 'The Understudy', year: 2023, runtime: '1h 31m',
      still: 'still-quiet', poster: 'p-understudy',
      short: 'The worst actor in the company gets one night on stage. He has not learned the lines.'
    },
    {
      id: 'static', title: 'Static', year: 2025, runtime: '1h 49m',
      still: 'still-vesper', poster: 'p-static',
      short: 'The old Scala shows something different every night, and nobody runs the projector.'
    }
  ];

  var byId = {};
  FILMS.forEach(function (f) { byId[f.id] = f; });

  var DEMO_EMAIL = 'neve@example.com';

  /* --------------------------------------------------------------- state */

  var state = { user: null, mode: 'signin', next: null };
  var teardown = [];

  /* --------------------------------------------------------------- utils */

  function esc(s) {
    return String(s).replace(/[&<>"]/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c];
    });
  }

  function ic(name, cls) {
    return '<svg class="ic' + (cls ? ' ' + cls : '') + '" aria-hidden="true" focusable="false"><use href="#i-' + name + '"/></svg>';
  }

  function goo(label, wide, link) {
    var inner =
      '<span class="goo-bg"><i class="body"></i><i class="knob"></i></span>' +
      '<span class="goo-fg"><span class="lbl">' + esc(label) + '</span>' +
      '<span class="knob">' + ic('arrow') + '</span></span>';
    if (link) return '<a class="goo" href="' + link + '">' + inner + '</a>';
    return '<button class="goo' + (wide ? ' wide' : '') + '" type="submit">' + inner + '</button>';
  }

  var SCREENS = {
    'catalogue': '/',
    'catalogue-signed-in': '/?u=1',
    'auth-signin': '/auth?next=halcyon',
    'auth-signin-error': '/auth?next=halcyon&err=bad&email=neve%40example.com',
    'auth-signup': '/auth?next=halcyon&mode=signup',
    'auth-signup-invalid': '/auth?next=halcyon&mode=signup&err=invalid&email=neve',
    'auth-signup-duplicate': '/auth?next=halcyon&mode=signup&err=dup&email=neve%40example.com',
    'auth-expired': '/auth?next=halcyon&err=expired',
    'player': '/play/halcyon?demo=1',
    'player-playing': '/play/halcyon?demo=1&playing=1',
    'player-idle': '/play/halcyon?demo=1&idle=1',
    'player-error': '/play/halcyon?demo=1&error=video',
    'player-missing': '/play/unlisted'
  };

  function parse(raw) {
    var qi = raw.indexOf('?');
    var path = qi < 0 ? raw : raw.slice(0, qi);
    return { path: path || '/', q: new URLSearchParams(qi < 0 ? '' : raw.slice(qi + 1)) };
  }

  function route() {
    return parse(location.hash.slice(1) || '/');
  }

  /* The preview opens a screen as index.html#<screen id>, the id in
     ui/mock/screens.json. Those ids route to the state they name; every other
     hash is one of the app's own routes. */
  function screenRoute() {
    var r = route();
    var target = Object.prototype.hasOwnProperty.call(SCREENS, r.path) ? SCREENS[r.path] : null;
    return target ? parse(target) : r;
  }

  function nav(path) {
    var target = '#' + path;
    if (location.hash === target) render();
    else location.hash = target;
  }

  function onTeardown(fn) { teardown.push(fn); }

  function initial(email) { return (email || '?').charAt(0).toUpperCase(); }

  function brand(href) {
    return '<a class="brand" href="' + href + '">' + ic('clover', 'ic-brand') + '<span>Clover</span></a>';
  }

  function header() {
    var links = '<nav class="hdr-nav" aria-label="Primary">' +
        '<button type="button" data-scroll="the-catalogue">Catalogue</button>' +
        '<button type="button" data-scroll="how-it-works">How it works</button>' +
      '</nav>';
    var right = state.user
      ? '<span class="who" title="Signed in as ' + esc(state.user.email) + '" aria-label="Signed in as ' + esc(state.user.email) + '">' +
          '<i aria-hidden="true">' + initial(state.user.email) + '</i></span>' +
        '<button class="hdr-link" type="button" id="signout">Sign out</button>'
      : goo('Sign in', false, '#/auth');
    return '<header class="hdr">' +
      '<div class="wrap hdr-in">' + brand('#/') + links +
        '<div class="hdr-right">' + right + '</div>' +
      '</div></header>';
  }

  /* ------------------------------------------------------------ catalogue */

  function pageCatalogue() {
    var cards = FILMS.map(function (f, i) {
      return '<li class="card" style="--i:' + Math.min(i, 6) + '">' +
        '<a class="card-a" href="#/play/' + f.id + '">' +
          '<span class="card-art"><img src="' + IMG + f.poster + '.jpg" alt="Poster for ' + esc(f.title) + '"' +
            (i < 5 ? ' fetchpriority="high"' : ' loading="lazy"') + '></span>' +
          '<span class="card-title">' + esc(f.title) + '</span>' +
          '<span class="card-desc">' + esc(f.short) + '</span>' +
        '</a></li>';
    }).join('');

    var rail = FILMS.slice(0, 2).map(function (f) {
      return '<a class="rail-item" href="#/play/' + f.id + '">' +
          '<span class="rail-thumb"><img src="' + IMG + f.poster + '.jpg" alt=""></span>' +
          '<span class="rail-chip">' + f.year + '</span>' +
          '<span class="rail-text">' +
            '<span class="rail-title">' + esc(f.title) + '</span>' +
            '<span class="rail-desc">' + esc(f.rail || f.short) + '</span>' +
          '</span>' +
          '<span class="rail-go">' + ic('arrow') + '</span>' +
        '</a>';
    }).join('');

    var steps = [
      ['eye', 'Browse', 'Every film, no account.'],
      ['lock', 'Sign in', 'One email, one password.'],
      ['play', 'Watch', 'Play, pause, seek, fullscreen.']
    ].map(function (s, i) {
      return '<div class="step" style="--i:' + i + '">' +
          '<span class="step-ic">' + ic(s[0]) + '</span>' +
          '<span class="step-body"><b>' + s[1] + '</b><p>' + s[2] + '</p></span>' +
        '</div>';
    }).join('');

    return header() +
      '<main id="main" tabindex="-1">' +
        '<section class="hero" aria-labelledby="hero-title">' +
          '<div class="hero-media">' +
            '<img src="' + IMG + 'still-vesper.jpg" alt="" fetchpriority="high">' +
          '</div>' +
          '<div class="wrap hero-in">' +
            '<div class="hero-copy">' +
              '<div class="hero-line"><h1 class="hero-title" id="hero-title">Pick a film.</h1></div>' +
              '<div class="hero-sub">' +
                '<p class="hero-lede">The whole catalogue is open. Watching takes an account, and Clover only asks when you press play.</p>' +
                '<button class="hero-ico" type="button" data-scroll="the-catalogue" aria-label="See all ' + FILMS.length + ' films">' + ic('down') + '</button>' +
              '</div>' +
            '</div>' +
            '<aside class="facts" aria-label="Catalogue at a glance">' +
              '<span class="fact">' + FILMS.length + ' films</span>' +
              '<span class="fact">2021&ndash;2025</span>' +
              '<button class="fact-go" type="button" data-scroll="the-catalogue" aria-label="See all ' + FILMS.length + ' films">' + ic('down') + '</button>' +
            '</aside>' +
            '<div class="rail">' + rail + '</div>' +
            '<div class="hero-foot">' +
              '<aside class="stat" aria-label="Clover at a glance">' +
                '<div class="stat-row"><b>' + FILMS.length + '</b><span>films in the catalogue, free to browse</span></div>' +
                '<div class="stat-row"><b>1</b><span>account is all watching takes</span></div>' +
              '</aside>' +
              '<div class="steps" id="how-it-works" tabindex="-1">' + steps + '</div>' +
            '</div>' +
          '</div>' +
        '</section>' +
        '<section class="cat wrap" id="the-catalogue" tabindex="-1" aria-labelledby="cat-title">' +
          '<div class="cat-head">' +
            '<h2 id="cat-title">The catalogue</h2>' +
            '<p>' + FILMS.length + ' films, free to browse</p>' +
          '</div>' +
          '<ul class="grid">' + cards + '</ul>' +
        '</section>' +
        '<footer class="foot wrap">' +
          '<b>Clover</b>' +
          '<span>A small cinema. The catalogue is open; watching takes an account.</span>' +
        '</footer>' +
      '</main>';
  }

  function mountCatalogue() {
    var out = document.getElementById('signout');
    if (out) {
      out.addEventListener('click', function () {
        state.user = null;
        nav('/');
      });
    }
  }

  /* In-page links are buttons, not hash links: the mock routes on the hash,
     so a "#section" link would re-render the page instead of scrolling it. */
  document.addEventListener('click', function (e) {
    var trigger = e.target.closest && e.target.closest('[data-scroll]');
    if (!trigger) return;
    var target = document.getElementById(trigger.getAttribute('data-scroll'));
    if (!target) return;
    var calm = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    target.scrollIntoView({ behavior: calm ? 'auto' : 'smooth', block: 'start' });
    target.focus({ preventScroll: true });
  });

  /* ----------------------------------------------------------------- auth */

  function authCopy(mode, film) {
    if (mode === 'signup') {
      return {
        title: 'Create an account',
        sub: film
          ? 'One email and a password, then the film starts.'
          : 'One email and a password. That is the whole form.'
      };
    }
    return {
      title: film ? 'Sign in to watch' : 'Sign in',
      sub: film ? 'Use your account and the film starts.' : 'Welcome back. The catalogue is waiting.'
    };
  }

  function pageAuth(q) {
    var nextId = q.get('next');
    var film = nextId ? byId[nextId] : null;
    var mode = q.get('mode') === 'signup' ? 'signup' : 'signin';
    var err = q.get('err') || '';
    state.mode = mode;
    state.next = film ? film.id : null;

    var copy = authCopy(mode, film);
    var email = q.get('email') || '';
    var pass = err === 'invalid' ? 'abc' : '';

    var notice = '';
    if (err === 'bad') {
      notice = '<div class="notice err" role="alert">' + ic('alert') +
        '<span>That email and password do not match an account.</span></div>';
    } else if (err === 'expired') {
      notice = '<div class="notice info" role="status">' + ic('lock') +
        '<span>Your session ended. Sign in again to keep watching.</span></div>';
    }

    var art = film
      ? '<div class="art-foot">' +
          '<h2 class="art-title">' + esc(film.title) + '</h2>' +
          '<p class="art-desc">' + esc(film.short) + '</p>' +
          '<p class="art-meta">' + film.year + ' &middot; ' + esc(film.runtime) + '</p>' +
        '</div>'
      : '<div class="art-foot">' +
          '<h2 class="art-title">Clover</h2>' +
          '<p class="art-desc">A small cinema. Every film in the catalogue, open to browse.</p>' +
          '<p class="art-meta">Sign in when you want to watch</p>' +
        '</div>';

    var still = film ? film.still : 'still-quiet';

    var emailErr = err === 'dup'
      ? '<p class="ferr" id="e-email">An account already uses this email. ' +
        '<button type="button" class="linkbtn" id="to-signin">Sign in instead.</button></p>'
      : err === 'invalid'
        ? '<p class="ferr" id="e-email">Enter an email address like you@example.com.</p>'
        : '<p class="ferr" id="e-email" hidden></p>';
    var passErr = err === 'invalid'
      ? '<p class="ferr" id="e-pass">Use at least 8 characters.</p>'
      : '<p class="ferr" id="e-pass" hidden></p>';

    return '<div class="auth">' +
        '<aside class="auth-art" aria-hidden="true">' +
          '<div class="art-media"><img src="' + IMG + still + '.jpg" alt=""></div>' +
          art +
        '</aside>' +
        '<div class="auth-top">' + brand('#/') +
          '<a class="back" href="#/">' + ic('back') + '<span>Back to the catalogue</span></a>' +
        '</div>' +
        '<main id="main" tabindex="-1" class="auth-main">' +
          '<div class="auth-box">' +
            '<h1 id="a-title">' + copy.title + '</h1>' +
            '<p class="auth-sub" id="a-sub">' + copy.sub + '</p>' +
            '<div class="tabs" role="group" aria-label="Sign in or create an account">' +
              '<button class="tab" type="button" data-mode="signin" aria-pressed="' + (mode === 'signin') + '">Sign in</button>' +
              '<button class="tab" type="button" data-mode="signup" aria-pressed="' + (mode === 'signup') + '">Create account</button>' +
            '</div>' +
            notice +
            '<form id="a-form" novalidate>' +
              '<div class="field' + (err === 'invalid' || err === 'dup' ? ' invalid' : '') + '" id="f-email">' +
                '<label for="email">Email</label>' +
                '<input id="email" name="email" type="email" autocomplete="email" placeholder="you@example.com" value="' + esc(email) + '">' +
                emailErr +
              '</div>' +
              '<div class="field' + (err === 'invalid' ? ' invalid' : '') + '" id="f-pass">' +
                '<label for="password">Password</label>' +
                '<input id="password" name="password" type="password" autocomplete="' + (mode === 'signup' ? 'new-password' : 'current-password') + '" value="' + esc(pass) + '">' +
                passErr +
                '<p class="fhint"' + (mode === 'signup' ? '' : ' hidden') + ' id="a-hint">At least 8 characters, including a letter and a number.</p>' +
              '</div>' +
              '<div class="auth-actions">' + goo(mode === 'signup' ? 'Create account' : 'Sign in', true) + '</div>' +
            '</form>' +
          '</div>' +
        '</main>' +
      '</div>';
  }

  function clearErrors() {
    document.querySelectorAll('.field').forEach(function (f) { f.classList.remove('invalid'); });
    document.querySelectorAll('.ferr').forEach(function (e) {
      e.hidden = true;
      e.textContent = '';
    });
  }

  function setMode(mode, film) {
    state.mode = mode;
    var copy = authCopy(mode, film);
    document.getElementById('a-title').textContent = copy.title;
    document.getElementById('a-sub').textContent = copy.sub;
    document.querySelectorAll('.tab').forEach(function (t) {
      t.setAttribute('aria-pressed', String(t.dataset.mode === mode));
    });
    var lbl = document.querySelector('.auth-actions .lbl');
    if (lbl) lbl.textContent = mode === 'signup' ? 'Create account' : 'Sign in';
    var hint = document.getElementById('a-hint');
    if (hint) hint.hidden = mode !== 'signup';
    document.getElementById('password').setAttribute('autocomplete', mode === 'signup' ? 'new-password' : 'current-password');
    clearErrors();
  }

  function fieldError(id, msg) {
    var field = document.getElementById('f-' + id);
    var err = document.getElementById('e-' + id);
    if (!field || !err) return;
    field.classList.add('invalid');
    err.textContent = msg;
    err.hidden = false;
  }

  function validate(mode, email, password) {
    var ok = true;
    if (!email.trim()) { fieldError('email', 'Enter your email address.'); ok = false; }
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      fieldError('email', 'Enter an email address like you@example.com.'); ok = false;
    }
    if (!password) { fieldError('pass', 'Enter your password.'); ok = false; }
    else if (mode === 'signup') {
      if (password.length < 8) { fieldError('pass', 'Use at least 8 characters.'); ok = false; }
      else if (!/[a-z]/i.test(password)) { fieldError('pass', 'Add at least one letter.'); ok = false; }
      else if (!/[0-9]/.test(password)) { fieldError('pass', 'Add at least one number.'); ok = false; }
    }
    return ok;
  }

  function mountAuth() {
    var film = state.next ? byId[state.next] : null;

    document.querySelectorAll('.tab').forEach(function (t) {
      t.addEventListener('click', function () { setMode(t.dataset.mode, film); });
    });

    var toSignin = document.getElementById('to-signin');
    if (toSignin) toSignin.addEventListener('click', function () { setMode('signin', film); });

    document.getElementById('a-form').addEventListener('submit', function (e) {
      e.preventDefault();
      clearErrors();
      var email = document.getElementById('email').value.trim();
      var pass = document.getElementById('password').value;
      if (!validate(state.mode, email, pass)) {
        var bad = document.querySelector('.field.invalid input');
        if (bad) bad.focus();
        return;
      }
      state.user = { email: email };
      nav(state.next ? '/play/' + state.next : '/');
    });
  }

  /* --------------------------------------------------------------- player */

  function fmt(s) {
    s = Math.max(0, Math.floor(s || 0));
    var h = Math.floor(s / 3600);
    var m = Math.floor((s % 3600) / 60);
    var ss = String(s % 60).padStart(2, '0');
    return (h ? h + ':' + String(m).padStart(2, '0') : m) + ':' + ss;
  }

  function playerChrome(f) {
    return '<div class="p-top">' +
        '<a class="p-back" href="#/" aria-label="Back to the catalogue">' + ic('back') + '</a>' +
        '<div class="p-meta"><h1>' + esc(f.title) + '</h1><p>' + f.year + ' &middot; ' + esc(f.runtime) + '</p></div>' +
      '</div>' +
      '<button class="p-big" type="button" id="bigplay" aria-label="Play">' + ic('play') + '</button>' +
      '<div class="p-bar">' +
        '<button class="p-btn" type="button" id="toggle" aria-label="Play">' + ic('play') + '</button>' +
        '<span class="p-time" id="elapsed">0:00</span>' +
        '<input class="p-scrub" id="scrub" type="range" min="0" max="14" step="0.05" value="0" aria-label="Seek">' +
        '<span class="p-time p-dur" id="duration">0:14</span>' +
        '<div class="p-vol">' +
          '<button class="p-btn" type="button" id="mute" aria-label="Mute">' + ic('volume') + '</button>' +
          '<input class="p-vrange" id="vrange" type="range" min="0" max="1" step="0.01" value="1" aria-label="Volume">' +
        '</div>' +
        '<button class="p-btn" type="button" id="full" aria-label="Enter fullscreen">' + ic('full') + '</button>' +
      '</div>';
  }

  function pagePlayer(f, q) {
    return '<main id="main" tabindex="-1" class="player">' +
        '<img class="p-still" id="p-still" src="' + IMG + f.still + '.jpg" alt="">' +
        '<video id="v" playsinline preload="metadata" poster="' + IMG + f.still + '.jpg">' +
          '<source src="' + CLIP + '" type="video/mp4">' +
        '</video>' +
        playerChrome(f) +
      '</main>';
  }

  function pagePlayerError(error, id) {
    var missing = error === 'missing';
    return '<main id="main" tabindex="-1" class="player is-msg">' +
      '<div class="p-msg">' +
        '<h1>' + (missing ? 'This film is not available.' : 'The film could not be loaded.') + '</h1>' +
        '<p>' + (missing
          ? 'It may have left the catalogue, or the link may be wrong.'
          : 'Check your connection, then try again.') + '</p>' +
        '<div class="p-msg-actions">' +
          (missing ? '' : goo('Try again', false, '#/play/' + id + '?demo=1')) +
          '<a class="link" href="#/">Back to the catalogue</a>' +
        '</div>' +
      '</div>' +
    '</main>';
  }

  function mountPlayer(f, q) {
    document.body.classList.add('in-player');
    var wrap = document.getElementById('main');
    var v = document.getElementById('v');
    var still = document.getElementById('p-still');
    var big = document.getElementById('bigplay');
    var toggle = document.getElementById('toggle');
    var scrub = document.getElementById('scrub');
    var elapsed = document.getElementById('elapsed');
    var duration = document.getElementById('duration');
    var muteBtn = document.getElementById('mute');
    var vrange = document.getElementById('vrange');
    var full = document.getElementById('full');
    var hideTimer = null;
    var scrubbing = false;
    var mutedByUs = false;
    var idleLocked = false;
    var holdChrome = q.get('playing') === '1';
    var live = true;
    var wantPlay = false;
    var clock = { t: 0, dur: 14, paused: true, volume: 1, muted: false, raf: 0, last: 0 };

    function getDur() { return live ? (v.duration || 14) : clock.dur; }
    function getTime() { return live ? (v.currentTime || 0) : clock.t; }
    function isPaused() { return live ? v.paused : clock.paused; }
    function volume() { return live ? v.volume : clock.volume; }
    function isMuted() { return (live ? v.muted : clock.muted) || volume() === 0; }

    function paint() {
      var d = getDur();
      var t = getTime();
      var pct = d ? (t / d) * 100 : 0;
      scrub.max = String(d);
      scrub.value = String(t);
      scrub.style.setProperty('--fill', pct + '%');
      scrub.setAttribute('aria-valuetext', fmt(t) + ' of ' + fmt(d));
      elapsed.textContent = fmt(t);
      duration.textContent = fmt(d);
    }

    function paintVolume() {
      var muted = isMuted();
      muteBtn.innerHTML = ic(muted ? 'mute' : 'volume');
      muteBtn.setAttribute('aria-label', muted ? 'Unmute' : 'Mute');
      vrange.style.setProperty('--fill', (muted ? 0 : volume()) * 100 + '%');
      vrange.value = String(muted ? 0 : volume());
    }

    function paintPlay() {
      var paused = isPaused();
      toggle.innerHTML = ic(paused ? 'play' : 'pause');
      toggle.setAttribute('aria-label', paused ? 'Play' : 'Pause');
      big.hidden = !paused;
      if (still) still.classList.toggle('drift', !paused);
    }

    function showChrome() {
      if (idleLocked) return;
      wrap.classList.remove('idle');
      clearTimeout(hideTimer);
      if (holdChrome) return;
      hideTimer = setTimeout(function () {
        if (!isPaused()) wrap.classList.add('idle');
      }, 2600);
    }

    /* /ui-preview's inner CSP allows no media, so the clip never loads there.
       The still and a clock stand in so every control stays demonstrable;
       opening the mock directly still plays the real file. */
    function fallback() {
      if (!live) return;
      live = false;
      var wasPlaying = wantPlay || !v.paused;
      clock.t = Math.min(clock.dur, v.currentTime || 0);
      clock.paused = true;
      clock.volume = v.volume || 1;
      clock.muted = v.muted;
      v.hidden = true;
      wrap.classList.add('no-media');
      paint();
      paintPlay();
      paintVolume();
      if (wasPlaying) doPlay();
    }

    function tick(ts) {
      if (clock.paused) return;
      if (clock.last) clock.t += Math.min(0.25, (ts - clock.last) / 1000);
      clock.last = ts;
      clock.t = Math.min(clock.dur, clock.t);
      paint();
      if (clock.t >= clock.dur) {
        clock.paused = true;
        clock.raf = 0;
        clock.last = 0;
        paintPlay();
        wrap.classList.remove('idle');
        return;
      }
      clock.raf = requestAnimationFrame(tick);
    }

    function doPlay() {
      wantPlay = true;
      if (!live) {
        if (clock.paused) {
          clock.paused = false;
          clock.last = 0;
          clock.raf = requestAnimationFrame(tick);
          paintPlay();
          showChrome();
        }
        return;
      }
      var p = v.play();
      if (p && p.catch) {
        p.catch(function () {
          v.muted = true;
          mutedByUs = true;
          paintVolume();
          v.play().catch(function () {});
        });
      }
    }

    function doPause() {
      wantPlay = false;
      if (!live) {
        clock.paused = true;
        cancelAnimationFrame(clock.raf);
        clock.raf = 0;
        clock.last = 0;
        paintPlay();
        wrap.classList.remove('idle');
        return;
      }
      v.pause();
    }

    function togglePlay() {
      if (isPaused()) doPlay(); else doPause();
    }

    function seekTo(x) {
      x = Math.max(0, Math.min(getDur(), x));
      if (live) { if (v.duration) v.currentTime = x; }
      else clock.t = x;
    }

    function setVolume(x) {
      x = Math.max(0, Math.min(1, x));
      if (live) {
        v.muted = false;
        v.volume = x;
      } else {
        clock.muted = false;
        clock.volume = x;
        paintVolume();
      }
    }

    function setMuted(on) {
      if (live) v.muted = on; else clock.muted = on;
      paintVolume();
    }

    function unlockAudio() {
      if (mutedByUs && live) {
        mutedByUs = false;
        v.muted = false;
        paintVolume();
      }
    }

    v.addEventListener('loadedmetadata', function () { if (live) paint(); });
    v.addEventListener('timeupdate', function () { if (live && !scrubbing) paint(); });
    v.addEventListener('play', function () { if (live) { paintPlay(); showChrome(); } });
    v.addEventListener('pause', function () { if (live) { paintPlay(); wrap.classList.remove('idle'); } });
    v.addEventListener('ended', function () { if (live) { paintPlay(); wrap.classList.remove('idle'); } });
    v.addEventListener('volumechange', function () { if (live) paintVolume(); });
    v.addEventListener('error', fallback);
    if (v.readyState === 0) setTimeout(function () { if (live && v.readyState === 0) fallback(); }, 2500);

    toggle.addEventListener('click', togglePlay);
    big.addEventListener('click', function () { unlockAudio(); togglePlay(); });
    v.addEventListener('click', togglePlay);
    if (still) still.addEventListener('click', togglePlay);

    scrub.addEventListener('input', function () {
      scrubbing = true;
      seekTo(Number(scrub.value));
      paint();
    });
    scrub.addEventListener('change', function () { scrubbing = false; });

    vrange.addEventListener('input', function () {
      mutedByUs = false;
      setVolume(Number(vrange.value));
    });
    muteBtn.addEventListener('click', function () {
      mutedByUs = false;
      if (isMuted()) {
        setMuted(false);
        if (volume() === 0) setVolume(1);
      } else {
        setMuted(true);
      }
    });

    full.addEventListener('click', function () {
      if (document.fullscreenElement) document.exitFullscreen();
      else if (wrap.requestFullscreen) wrap.requestFullscreen().catch(function () {});
    });
    document.addEventListener('fullscreenchange', function () {
      var on = !!document.fullscreenElement;
      full.innerHTML = ic(on ? 'exit' : 'full');
      full.setAttribute('aria-label', on ? 'Exit fullscreen' : 'Enter fullscreen');
    });

    function onKey(e) {
      var tag = (e.target.tagName || '').toLowerCase();
      if (tag === 'input' || tag === 'button' || tag === 'a') return;
      if (e.key === ' ' || e.key === 'k') { e.preventDefault(); unlockAudio(); togglePlay(); }
      else if (e.key === 'ArrowRight') { e.preventDefault(); seekTo(getTime() + 5); paint(); }
      else if (e.key === 'ArrowLeft') { e.preventDefault(); seekTo(getTime() - 5); paint(); }
      else if (e.key === 'ArrowUp') { e.preventDefault(); setVolume(volume() + 0.1); }
      else if (e.key === 'ArrowDown') { e.preventDefault(); setVolume(volume() - 0.1); }
      else if (e.key === 'm') { muteBtn.click(); }
      else if (e.key === 'f') { full.click(); }
      showChrome();
    }

    function onMove() {
      idleLocked = false;
      holdChrome = false;
      showChrome();
    }

    document.addEventListener('keydown', onKey);
    wrap.addEventListener('pointermove', onMove);
    wrap.addEventListener('pointerdown', unlockAudio);

    paint();
    paintPlay();
    paintVolume();

    if (q.get('paused') === '1') doPause();
    if (q.get('playing') === '1' || q.get('idle') === '1') doPlay();
    if (q.get('idle') === '1') {
      idleLocked = true;
      setTimeout(function () { wrap.classList.add('idle'); }, 120);
    }

    onTeardown(function () {
      clearTimeout(hideTimer);
      cancelAnimationFrame(clock.raf);
      document.removeEventListener('keydown', onKey);
      v.pause();
    });
  }

  /* ------------------------------------------------------------------ app */

  function render() {
    teardown.forEach(function (fn) { try { fn(); } catch (e) {} });
    teardown = [];

    var r = screenRoute();
    var app = document.getElementById('app');
    var isPlayer = r.path.indexOf('/play/') === 0;

    document.body.classList.remove('in-player');
    document.title = 'Clover — pick a film';

    if (r.path === '/' || r.path === '') {
      if (r.q.get('u') === '1') state.user = { email: DEMO_EMAIL };
      app.innerHTML = pageCatalogue();
      mountCatalogue();
      document.title = 'Clover — the catalogue';
    } else if (r.path === '/auth') {
      app.innerHTML = pageAuth(r.q);
      mountAuth();
      document.title = state.mode === 'signup' ? 'Clover — create an account' : 'Clover — sign in';
    } else if (isPlayer) {
      var id = r.path.slice('/play/'.length);
      var f = byId[id];
      if (r.q.get('demo') === '1') state.user = state.user || { email: DEMO_EMAIL };
      if (!f) {
        app.innerHTML = pagePlayerError('missing');
        document.body.classList.add('in-player');
        document.title = 'Clover — film not available';
      } else if (r.q.get('error') === 'video') {
        app.innerHTML = pagePlayerError('video', f.id);
        document.body.classList.add('in-player');
        document.title = 'Clover — ' + f.title;
      } else if (!state.user) {
        location.replace('#/auth?next=' + encodeURIComponent(f.id));
        return;
      } else {
        app.innerHTML = pagePlayer(f, r.q);
        mountPlayer(f, r.q);
        document.title = 'Clover — ' + f.title;
      }
    } else {
      nav('/');
      return;
    }

    if (!isPlayer) window.scrollTo(0, 0);
    var main = document.getElementById('main');
    if (main) main.focus({ preventScroll: true });
  }

  window.addEventListener('hashchange', render);

  var lastY = -1;
  window.addEventListener('scroll', function () {
    var y = window.scrollY;
    if (y === lastY) return;
    lastY = y;
    var h = document.querySelector('.hdr');
    if (h) h.classList.toggle('solid', y > 16);
  }, { passive: true });

  render();
}());
