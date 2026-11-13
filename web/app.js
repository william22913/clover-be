(function () {
  'use strict';

  /* ---------------------------------------------------------------- utils */

  function esc(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }

  function ic(name, cls) {
    return '<svg class="ic' + (cls ? ' ' + cls : '') + '" aria-hidden="true" focusable="false"><use href="#i-' + name + '"/></svg>';
  }

  function goo(label, wide, href) {
    var inner =
      '<span class="goo-bg"><i class="body"></i><i class="knob"></i></span>' +
      '<span class="goo-fg"><span class="lbl">' + esc(label) + '</span>' +
      '<span class="knob">' + ic('arrow') + '</span></span>';
    if (href) return '<a class="goo" href="' + href + '">' + inner + '</a>';
    return '<button class="goo' + (wide ? ' wide' : '') + '" type="submit">' + inner + '</button>';
  }

  function el(id) { return document.getElementById(id); }

  function initial(email) { return (email || '?').charAt(0).toUpperCase(); }

  /* Playback clock format: m:ss or h:mm:ss. */
  function clock(s) {
    s = Math.max(0, Math.floor(s || 0));
    var h = Math.floor(s / 3600);
    var m = Math.floor((s % 3600) / 60);
    var ss = String(s % 60).padStart(2, '0');
    return (h ? h + ':' + String(m).padStart(2, '0') : m) + ':' + ss;
  }

  /* Catalogue duration format, the mock's own: 2h 06m. */
  function duration(s) {
    var h = Math.floor((s || 0) / 3600);
    var m = Math.round(((s || 0) % 3600) / 60);
    return h ? h + 'h ' + String(m).padStart(2, '0') + 'm' : m + 'm';
  }

  /* ---------------------------------------------------------------- state */

  var state = { user: null, films: [], mode: 'signin', next: null, booted: false };
  var teardown = [];
  var seq = 0;

  function onTeardown(fn) { teardown.push(fn); }
  function clearTeardown() {
    teardown.forEach(function (fn) { try { fn(); } catch (e) {} });
    teardown = [];
  }

  /* ------------------------------------------------------------------ api */

  function api(path, opts) {
    opts = opts || {};
    var init = {
      method: opts.method || 'GET',
      credentials: 'same-origin',
      headers: { 'Accept': 'application/json' }
    };
    if (opts.body !== undefined) {
      init.headers['Content-Type'] = 'application/json';
      init.body = JSON.stringify(opts.body);
    }
    return fetch(path, init).then(function (res) {
      if (res.status === 204) return null;
      return res.json().catch(function () { return null; }).then(function (data) {
        if (res.ok) return data;
        var raw = (data && data.error) || {};
        var err = new Error(raw.message || 'Something went wrong. Please try again.');
        err.code = raw.code || 'internal_error';
        err.fields = raw.fields || {};
        err.requestId = raw.request_id || '';
        err.status = res.status;
        throw err;
      });
    });
  }

  /* -------------------------------------------------------------- routing */

  function current() {
    return {
      path: location.pathname.replace(/\/+$/, '') || '/',
      q: new URLSearchParams(location.search)
    };
  }

  function navigate(path, replace) {
    var url = new URL(path, location.origin);
    var target = url.pathname + url.search;
    if (replace) history.replaceState(null, '', target);
    else history.pushState(null, '', target);
    render();
  }

  document.addEventListener('click', function (e) {
    var trigger = e.target.closest && e.target.closest('[data-scroll]');
    if (trigger) {
      var target = document.getElementById(trigger.getAttribute('data-scroll'));
      if (target) {
        var calm = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
        target.scrollIntoView({ behavior: calm ? 'auto' : 'smooth', block: 'start' });
        target.focus({ preventScroll: true });
      }
      return;
    }
    var a = e.target.closest && e.target.closest('a[href]');
    if (!a || a.target || a.origin !== location.origin) return;
    var href = a.getAttribute('href');
    if (href.charAt(0) !== '/') return;
    e.preventDefault();
    navigate(a.pathname + a.search);
  });

  window.addEventListener('popstate', function () { render(); });

  var lastY = -1;
  window.addEventListener('scroll', function () {
    var y = window.scrollY;
    if (y === lastY) return;
    lastY = y;
    var h = document.querySelector('.hdr');
    if (h) h.classList.toggle('solid', y > 16);
  }, { passive: true });

  /* --------------------------------------------------------------- header */

  function brand() {
    return '<a class="brand" href="/">' + ic('clover', 'ic-brand') + '<span>Clover</span></a>';
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
      : goo('Sign in', false, '/auth');
    return '<header class="hdr">' +
      '<div class="wrap hdr-in">' + brand() + links +
        '<div class="hdr-right">' + right + '</div>' +
      '</div></header>';
  }

  function mountHeader() {
    var out = el('signout');
    if (!out) return;
    out.addEventListener('click', function () {
      api('/v1/auth/sign-out', { method: 'POST' }).catch(function () {}).then(function () {
        state.user = null;
        navigate('/');
      });
    });
  }

  /* ------------------------------------------------------------- catalogue */

  function filmModel(m) {
    return {
      slug: m.slug,
      title: m.title,
      short: m.short_description,
      year: m.release_year,
      runtime: duration(m.runtime_seconds),
      poster: m.poster_url,
      still: m.still_url || m.poster_url
    };
  }

  function catalogueHTML(films) {
    var cards = films.map(function (f, i) {
      return '<li class="card">' +
        '<a class="card-a" href="/play/' + encodeURIComponent(f.slug) + '">' +
          '<span class="card-art"><img src="' + esc(f.poster) + '" alt="Poster for ' + esc(f.title) + '"' +
            (i < 5 ? ' fetchpriority="high"' : ' loading="lazy"') + '></span>' +
          '<span class="card-title">' + esc(f.title) + '</span>' +
          '<span class="card-desc">' + esc(f.short) + '</span>' +
        '</a></li>';
    }).join('');

    var rail = films.slice(0, 2).map(function (f) {
      return '<a class="rail-item" href="/play/' + encodeURIComponent(f.slug) + '">' +
          '<span class="rail-thumb"><img src="' + esc(f.poster) + '" alt=""></span>' +
          '<span class="rail-chip">' + f.year + '</span>' +
          '<span class="rail-text">' +
            '<span class="rail-title">' + esc(f.title) + '</span>' +
            '<span class="rail-desc">' + esc(f.short) + '</span>' +
          '</span>' +
          '<span class="rail-go">' + ic('arrow') + '</span>' +
        '</a>';
    }).join('');

    var steps = [
      ['eye', 'Browse', 'Every film, no account.'],
      ['lock', 'Sign in', 'One email, one password.'],
      ['play', 'Watch', 'Play, pause, seek, fullscreen.']
    ].map(function (s) {
      return '<div class="step">' +
          '<span class="step-ic">' + ic(s[0]) + '</span>' +
          '<span class="step-body"><b>' + s[1] + '</b><p>' + s[2] + '</p></span>' +
        '</div>';
    }).join('');

    return header() +
      '<main id="main" tabindex="-1">' +
        '<section class="hero" aria-labelledby="hero-title">' +
          '<div class="hero-media"><img src="/media/still/still-vesper.jpg" alt="" fetchpriority="high"></div>' +
          '<div class="wrap hero-in">' +
            '<div class="hero-copy">' +
              '<div class="hero-line"><h1 class="hero-title" id="hero-title">Pick a film.</h1></div>' +
              '<div class="hero-sub">' +
                '<p class="hero-lede">The whole catalogue is open. Watching takes an account, and Clover only asks when you press play.</p>' +
                '<button class="hero-ico" type="button" data-scroll="the-catalogue" aria-label="See all ' + films.length + ' films">' + ic('down') + '</button>' +
              '</div>' +
            '</div>' +
            '<aside class="facts" aria-label="Catalogue at a glance">' +
              '<span class="fact">' + films.length + ' films</span>' +
              '<span class="fact">2021&ndash;2025</span>' +
              '<button class="fact-go" type="button" data-scroll="the-catalogue" aria-label="See all ' + films.length + ' films">' + ic('down') + '</button>' +
            '</aside>' +
            '<div class="rail">' + rail + '</div>' +
            '<div class="hero-foot">' +
              '<aside class="stat" aria-label="Clover at a glance">' +
                '<div class="stat-row"><b>' + films.length + '</b><span>films in the catalogue, free to browse</span></div>' +
                '<div class="stat-row"><b>1</b><span>account is all watching takes</span></div>' +
              '</aside>' +
              '<div class="steps" id="how-it-works" tabindex="-1">' + steps + '</div>' +
            '</div>' +
          '</div>' +
        '</section>' +
        '<section class="cat wrap" id="the-catalogue" tabindex="-1" aria-labelledby="cat-title">' +
          '<div class="cat-head">' +
            '<h2 id="cat-title">The catalogue</h2>' +
            '<p>' + films.length + ' films, free to browse</p>' +
          '</div>' +
          '<ul class="grid">' + cards + '</ul>' +
        '</section>' +
        '<footer class="foot wrap">' +
          '<b>Clover</b>' +
          '<span>A small cinema. The catalogue is open; watching takes an account.</span>' +
        '</footer>' +
      '</main>';
  }

  function pageCatalogue(my, q) {
    document.title = 'Clover — pick a film';
    var app = el('app');
    app.innerHTML = header() + '<div class="loading">Loading the catalogue&hellip;</div>';
    mountHeader();

    return api('/v1/movies').then(function (data) {
      if (my !== seq) return;
      state.films = data.items.map(filmModel);
      app.innerHTML = catalogueHTML(state.films);
      mountHeader();
      document.title = 'Clover — pick a film';
      window.scrollTo(0, 0);
      var main = el('main');
      if (main) main.focus({ preventScroll: true });
    }).catch(function () {
      if (my !== seq) return;
      app.innerHTML = header() + routeMessage(
        'The catalogue could not be loaded.',
        'Check your connection, then try again.',
        'reload', 'Try again');
      mountHeader();
      var retry = el('reload');
      if (retry) retry.addEventListener('click', function () { render(); });
    });
  }

  /* ---------------------------------------------------------------- shared */

  function routeMessage(title, body, buttonID, buttonLabel, withBack) {
    return '<main class="route-msg"><div class="p-msg">' +
        '<h1>' + esc(title) + '</h1>' +
        '<p>' + esc(body) + '</p>' +
        '<div class="p-msg-actions">' +
          (buttonID ? goo(buttonLabel, false, null).replace('<button class="goo"', '<button id="' + buttonID + '" class="goo"') : '') +
          (withBack ? '<a class="link" href="/">Back to the catalogue</a>' : '') +
        '</div>' +
      '</div></main>';
  }

  function notice(kind, message) {
    var wrap = el('a-notice');
    if (!wrap) return;
    wrap.innerHTML = '<div class="notice ' + kind + '" role="alert">' +
      ic(kind === 'info' ? 'lock' : 'alert') +
      '<span>' + esc(message) + '</span></div>';
    wrap.hidden = false;
  }

  /* ----------------------------------------------------------------- auth */

  function authCopy(mode, film) {
    if (mode === 'signup') {
      return {
        title: 'Create an account',
        sub: film ? 'One email and a password, then the film starts.' : 'One email and a password. That is the whole form.'
      };
    }
    return {
      title: film ? 'Sign in to watch' : 'Sign in',
      sub: film ? 'Use your account and the film starts.' : 'Welcome back. The catalogue is waiting.'
    };
  }

  function authHTML(mode, film, reason) {
    var copy = authCopy(mode, film);
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

    var still = film ? film.still : '/media/still/still-quiet.jpg';
    var expired = reason === 'expired'
      ? '<div class="notice info" role="status">' + ic('lock') +
        '<span>Your session ended. Sign in again to keep watching.</span></div>'
      : '';

    return '<div class="auth">' +
        '<aside class="auth-art" aria-hidden="true">' +
          '<div class="art-media"><img src="' + esc(still) + '" alt=""></div>' + art +
        '</aside>' +
        '<div class="auth-top">' + brand() +
          '<a class="back" href="/">' + ic('back') + '<span>Back to the catalogue</span></a>' +
        '</div>' +
        '<main id="main" tabindex="-1" class="auth-main">' +
          '<div class="auth-box">' +
            '<h1 id="a-title">' + esc(copy.title) + '</h1>' +
            '<p class="auth-sub" id="a-sub">' + esc(copy.sub) + '</p>' +
            '<div class="tabs" role="group" aria-label="Sign in or create an account">' +
              '<button class="tab" type="button" data-mode="signin" aria-pressed="' + (mode === 'signin') + '">Sign in</button>' +
              '<button class="tab" type="button" data-mode="signup" aria-pressed="' + (mode === 'signup') + '">Create account</button>' +
            '</div>' +
            '<div id="a-notice">' + expired + '</div>' +
            '<form id="a-form" novalidate>' +
              '<div class="field" id="f-email">' +
                '<label for="email">Email</label>' +
                '<input id="email" name="email" type="email" autocomplete="email" placeholder="you@example.com">' +
                '<p class="ferr" id="e-email" hidden></p>' +
              '</div>' +
              '<div class="field" id="f-pass">' +
                '<label for="password">Password</label>' +
                '<input id="password" name="password" type="password" autocomplete="' + (mode === 'signup' ? 'new-password' : 'current-password') + '">' +
                '<p class="ferr" id="e-pass" hidden></p>' +
                '<p class="fhint" id="a-hint"' + (mode === 'signup' ? '' : ' hidden') + '>At least 8 characters, including a letter and a number.</p>' +
              '</div>' +
              '<div class="auth-actions">' + goo(mode === 'signup' ? 'Create account' : 'Sign in', true) + '</div>' +
            '</form>' +
          '</div>' +
        '</main>' +
      '</div>';
  }

  function pageAuth(my, q) {
    var rawNext = q.get('next') || '';
    var mode = q.get('mode') === 'signup' ? 'signup' : 'signin';
    var reason = q.get('reason') || '';
    var validNext = /^[a-z0-9-]{1,64}$/.test(rawNext) ? rawNext : '';
    var app = el('app');

    function done(film) {
      if (my !== seq) return;
      state.mode = mode;
      state.next = film ? film.slug : null;
      app.innerHTML = authHTML(mode, film, reason);
      mountAuth(film);
      document.title = mode === 'signup' ? 'Clover — create an account' : 'Clover — sign in';
      var main = el('main');
      if (main) main.focus({ preventScroll: true });
    }

    if (!validNext) { done(null); return null; }
    return api('/v1/movies/' + encodeURIComponent(validNext))
      .then(function (m) { done(filmModel(m)); })
      .catch(function () { done(null); });
  }

  function clearErrors() {
    document.querySelectorAll('.field').forEach(function (f) { f.classList.remove('invalid'); });
    document.querySelectorAll('.ferr').forEach(function (e) {
      e.hidden = true;
      e.textContent = '';
    });
  }

  function setFieldError(name, message) {
    var field = el('f-' + name);
    var err = el('e-' + name);
    if (!field || !err) return;
    field.classList.add('invalid');
    err.textContent = message;
    err.hidden = false;
  }

  function showFieldErrors(fields) {
    Object.keys(fields || {}).forEach(function (name) {
      setFieldError(name === 'password' ? 'pass' : name, fields[name]);
    });
    var bad = document.querySelector('.field.invalid input');
    if (bad) bad.focus();
  }

  function setMode(mode, film) {
    state.mode = mode;
    var copy = authCopy(mode, film);
    el('a-title').textContent = copy.title;
    el('a-sub').textContent = copy.sub;
    document.querySelectorAll('.tab').forEach(function (t) {
      t.setAttribute('aria-pressed', String(t.dataset.mode === mode));
    });
    var lbl = document.querySelector('.auth-actions .lbl');
    if (lbl) lbl.textContent = mode === 'signup' ? 'Create account' : 'Sign in';
    var hint = el('a-hint');
    if (hint) hint.hidden = mode !== 'signup';
    el('password').setAttribute('autocomplete', mode === 'signup' ? 'new-password' : 'current-password');
    var noticeBox = el('a-notice');
    if (noticeBox) noticeBox.innerHTML = '';
    clearErrors();
  }

  function clientValidation(mode, email, password) {
    var fields = {};
    if (!email) fields.email = 'Enter your email address.';
    else if (email.length > 254) fields.email = 'That email address is too long.';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) fields.email = 'Enter a valid email address.';

    if (!password) {
      fields.password = mode === 'signup' ? 'Enter a password.' : 'Enter your password.';
    } else if (mode === 'signup') {
      if (password.length < 8) fields.password = 'Use at least 8 characters.';
      else if (password.length > 72) fields.password = 'Use no more than 72 characters.';
      else if (!/[a-zA-Z]/.test(password)) fields.password = 'Include at least one letter.';
      else if (!/[0-9]/.test(password)) fields.password = 'Include at least one number.';
    }
    return fields;
  }

  function mountAuth(film) {
    document.querySelectorAll('.tab').forEach(function (t) {
      t.addEventListener('click', function () { setMode(t.dataset.mode, film); });
    });

    var form = el('a-form');
    form.addEventListener('submit', function (e) {
      e.preventDefault();
      clearErrors();
      var email = el('email').value.trim();
      var password = el('password').value;
      var fields = clientValidation(state.mode, email, password);
      if (Object.keys(fields).length) { showFieldErrors(fields); return; }

      var submit = form.querySelector('button[type="submit"]');
      submit.disabled = true;
      var path = state.mode === 'signup' ? '/v1/auth/sign-up' : '/v1/auth/sign-in';
      api(path, { method: 'POST', body: { email: email, password: password } })
        .then(function (data) {
          state.user = data.user;
          navigate(state.next ? '/play/' + state.next : '/');
        })
        .catch(function (err) {
          submit.disabled = false;
          handleAuthError(err, film);
        });
    });
  }

  function handleAuthError(err, film) {
    if (err.code === 'validation_failed') {
      var mapped = {};
      Object.keys(err.fields || {}).forEach(function (k) { mapped[k] = err.fields[k]; });
      showFieldErrors(mapped);
      return;
    }
    if (err.code === 'email_taken') {
      var field = el('f-email');
      var msg = el('e-email');
      if (field && msg) {
        field.classList.add('invalid');
        msg.textContent = 'An account already uses this email. ';
        var toSignIn = document.createElement('button');
        toSignIn.type = 'button';
        toSignIn.className = 'linkbtn';
        toSignIn.textContent = 'Sign in instead.';
        toSignIn.addEventListener('click', function () { setMode('signin', film); });
        msg.appendChild(toSignIn);
        msg.hidden = false;
        el('email').focus();
      }
      return;
    }
    if (err.code === 'invalid_credentials' || err.code === 'rate_limited' || err.code === 'session_expired') {
      notice('err', err.message);
      var pass = el('password');
      if (pass) pass.value = '';
      return;
    }
    notice('err', 'Something went wrong. Please try again.' + (err.requestId ? ' (Request ' + err.requestId + ')' : ''));
  }

  /* --------------------------------------------------------------- player */

  function playerHTML(f) {
    return '<main id="main" tabindex="-1" class="player">' +
        '<img class="p-still" id="p-still" src="' + esc(f.still) + '" alt="">' +
        '<video id="v" playsinline preload="metadata" poster="' + esc(f.still) + '">' +
          '<source src="/v1/movies/' + encodeURIComponent(f.slug) + '/video" type="video/mp4">' +
        '</video>' +
        '<div class="p-top">' +
          '<a class="p-back" href="/" aria-label="Back to the catalogue">' + ic('back') + '</a>' +
          '<div class="p-meta"><h1>' + esc(f.title) + '</h1><p>' + f.year + ' &middot; ' + esc(f.runtime) + '</p></div>' +
        '</div>' +
        '<button class="p-big" type="button" id="bigplay" aria-label="Play">' + ic('play') + '</button>' +
        '<div class="p-bar">' +
          '<button class="p-btn" type="button" id="toggle" aria-label="Play">' + ic('play') + '</button>' +
          '<span class="p-time" id="elapsed">0:00</span>' +
          '<input class="p-scrub" id="scrub" type="range" min="0" max="0" step="0.05" value="0" aria-label="Seek">' +
          '<span class="p-time p-dur" id="duration">0:00</span>' +
          '<div class="p-vol">' +
            '<button class="p-btn" type="button" id="mute" aria-label="Mute">' + ic('volume') + '</button>' +
            '<input class="p-vrange" id="vrange" type="range" min="0" max="1" step="0.01" value="1" aria-label="Volume">' +
          '</div>' +
          '<button class="p-btn" type="button" id="full" aria-label="Enter fullscreen">' + ic('full') + '</button>' +
        '</div>' +
      '</main>';
  }

  function playerMissingHTML() {
    return '<main class="player is-msg"><div class="p-msg">' +
        '<h1>This film is not available.</h1>' +
        '<p>It may have left the catalogue, or the link may be wrong.</p>' +
        '<div class="p-msg-actions"><a class="link" href="/">Back to the catalogue</a></div>' +
      '</div></main>';
  }

  function playerVideoErrorHTML(f) {
    return '<main class="player is-msg"><div class="p-msg">' +
        '<h1>The film could not be loaded.</h1>' +
        '<p>Check your connection, then try again.</p>' +
        '<div class="p-msg-actions">' +
          goo('Try again', false, null).replace('<button class="goo"', '<button id="retry" class="goo"') +
          '<a class="link" href="/">Back to the catalogue</a>' +
        '</div>' +
      '</div></main>';
  }

  function playerLoadErrorHTML() {
    return '<main class="player is-msg"><div class="p-msg">' +
        '<h1>The film could not be loaded.</h1>' +
        '<p>Check your connection, then try again.</p>' +
        '<div class="p-msg-actions">' +
          goo('Try again', false, null).replace('<button class="goo"', '<button id="retry" class="goo"') +
          '<a class="link" href="/">Back to the catalogue</a>' +
        '</div>' +
      '</div></main>';
  }

  function pagePlayer(my) {
    document.body.classList.add('in-player');
    var slug = decodeURIComponent(current().path.slice('/play/'.length));

    if (!state.user) {
      navigate('/auth?next=' + encodeURIComponent(slug), true);
      return null;
    }

    var app = el('app');
    app.innerHTML = '<main class="player is-msg"><div class="p-msg"><h1>Loading&hellip;</h1></div></main>';

    return api('/v1/movies/' + encodeURIComponent(slug))
      .then(function (movie) {
        if (my !== seq) return;
        var f = filmModel(movie);
        app.innerHTML = playerHTML(f);
        document.title = 'Clover — ' + f.title;
        mountPlayer(f);
      })
      .catch(function (err) {
        if (my !== seq) return;
        app.innerHTML = err.status === 404 ? playerMissingHTML() : playerLoadErrorHTML();
        document.title = 'Clover — film not available';
        var retry = el('retry');
        if (retry) retry.addEventListener('click', function () { navigate('/play/' + encodeURIComponent(slug), true); });
      });
  }

  function mountPlayer(f) {
    var wrap = el('main');
    var v = el('v');
    var still = el('p-still');
    var big = el('bigplay');
    var toggle = el('toggle');
    var scrub = el('scrub');
    var elapsed = el('elapsed');
    var durationEl = el('duration');
    var muteBtn = el('mute');
    var vrange = el('vrange');
    var full = el('full');
    var hideTimer = null;
    var scrubbing = false;
    var mutedByUs = false;
    var wantPlay = false;

    function paint() {
      var d = v.duration || 0;
      var t = v.currentTime || 0;
      var pct = d ? (t / d) * 100 : 0;
      scrub.max = String(d);
      scrub.value = String(t);
      scrub.style.setProperty('--fill', pct + '%');
      scrub.setAttribute('aria-valuetext', clock(t) + ' of ' + clock(d));
      elapsed.textContent = clock(t);
      durationEl.textContent = clock(d);
    }

    function paintVolume() {
      var muted = v.muted || v.volume === 0;
      muteBtn.innerHTML = ic(muted ? 'mute' : 'volume');
      muteBtn.setAttribute('aria-label', muted ? 'Unmute' : 'Mute');
      vrange.style.setProperty('--fill', (muted ? 0 : v.volume) * 100 + '%');
      vrange.value = String(muted ? 0 : v.volume);
    }

    function paintPlay() {
      var paused = v.paused;
      toggle.innerHTML = ic(paused ? 'play' : 'pause');
      toggle.setAttribute('aria-label', paused ? 'Play' : 'Pause');
      big.hidden = !paused;
      if (still) still.classList.toggle('drift', !paused);
    }

    function showChrome() {
      wrap.classList.remove('idle');
      clearTimeout(hideTimer);
      hideTimer = setTimeout(function () {
        if (!v.paused) wrap.classList.add('idle');
      }, 2600);
    }

    function doPlay() {
      wantPlay = true;
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
      v.pause();
    }

    function togglePlay() { if (v.paused) doPlay(); else doPause(); }

    function seekTo(x) {
      if (!v.duration) return;
      v.currentTime = Math.max(0, Math.min(v.duration, x));
    }

    function setVolume(x) {
      x = Math.max(0, Math.min(1, x));
      v.muted = false;
      v.volume = x;
    }

    function setMuted(on) {
      v.muted = on;
      paintVolume();
    }

    function unlockAudio() {
      if (mutedByUs) {
        mutedByUs = false;
        v.muted = false;
        paintVolume();
      }
    }

    function showVideoError() {
      clearTeardown();
      el('app').innerHTML = playerVideoErrorHTML(f);
      document.title = 'Clover — ' + f.title;
      var retry = el('retry');
      if (retry) retry.addEventListener('click', function () {
        navigate('/play/' + encodeURIComponent(f.slug), true);
      });
    }

    v.addEventListener('loadedmetadata', paint);
    v.addEventListener('timeupdate', function () { if (!scrubbing) paint(); });
    v.addEventListener('play', function () { paintPlay(); showChrome(); });
    v.addEventListener('pause', function () { paintPlay(); wrap.classList.remove('idle'); });
    v.addEventListener('ended', function () { paintPlay(); wrap.classList.remove('idle'); });
    v.addEventListener('volumechange', paintVolume);
    v.addEventListener('error', function () {
      api('/v1/auth/me').then(function () {
        showVideoError();
      }).catch(function (err) {
        if (err.status === 401) {
          state.user = null;
          navigate('/auth?next=' + encodeURIComponent(f.slug) + '&reason=expired', true);
          return;
        }
        showVideoError();
      });
    });

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
      if (v.muted || v.volume === 0) {
        setMuted(false);
        if (v.volume === 0) setVolume(1);
      } else {
        setMuted(true);
      }
    });

    full.addEventListener('click', function () {
      if (document.fullscreenElement) document.exitFullscreen();
      else if (wrap.requestFullscreen) wrap.requestFullscreen().catch(function () {});
    });
    function onFullscreen() {
      var on = !!document.fullscreenElement;
      full.innerHTML = ic(on ? 'exit' : 'full');
      full.setAttribute('aria-label', on ? 'Exit fullscreen' : 'Enter fullscreen');
    }
    document.addEventListener('fullscreenchange', onFullscreen);

    function onKey(e) {
      var tag = (e.target.tagName || '').toLowerCase();
      if (tag === 'input' || tag === 'button' || tag === 'a') return;
      if (e.key === ' ' || e.key === 'k') { e.preventDefault(); unlockAudio(); togglePlay(); }
      else if (e.key === 'ArrowRight') { e.preventDefault(); seekTo((v.currentTime || 0) + 5); paint(); }
      else if (e.key === 'ArrowLeft') { e.preventDefault(); seekTo((v.currentTime || 0) - 5); paint(); }
      else if (e.key === 'ArrowUp') { e.preventDefault(); setVolume(v.volume + 0.1); }
      else if (e.key === 'ArrowDown') { e.preventDefault(); setVolume(v.volume - 0.1); }
      else if (e.key === 'm') { muteBtn.click(); }
      else if (e.key === 'f') { full.click(); }
      showChrome();
    }

    function onMove() { showChrome(); }

    document.addEventListener('keydown', onKey);
    wrap.addEventListener('pointermove', onMove);
    wrap.addEventListener('pointerdown', unlockAudio);

    paint();
    paintPlay();
    paintVolume();
    showChrome();

    onTeardown(function () {
      clearTimeout(hideTimer);
      document.removeEventListener('keydown', onKey);
      document.removeEventListener('fullscreenchange', onFullscreen);
      v.pause();
    });
  }

  /* ------------------------------------------------------------------ app */

  function render() {
    clearTeardown();
    document.body.classList.remove('in-player');
    var my = ++seq;
    var r = current();

    if (r.path === '/') { pageCatalogue(my, r.q); return; }
    if (r.path === '/auth') { pageAuth(my, r.q); return; }
    if (r.path.indexOf('/play/') === 0) { pagePlayer(my); return; }
    pageCatalogue(my, r.q);
  }

  function boot() {
    api('/v1/auth/me')
      .then(function (data) { state.user = data.user; })
      .catch(function () { state.user = null; })
      .then(function () {
        state.booted = true;
        render();
      });
  }

  boot();
}());
