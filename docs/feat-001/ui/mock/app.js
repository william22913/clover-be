/* Clover design mock — hash-routed states and interactions.
   Screens are reviewable directly through #<id>; the normal journey is:
   catalogue -> pick a film -> sign-in (film kept) -> player. */

const FILMS = [
  { id: 'amber-drift', title: 'Amber Drift', desc: 'A slow descent into a golden world where every surface is alive.' },
  { id: 'motel-sun', title: 'Motel Sun', desc: 'A stranded traveler waits out the night in a motel that should have closed years ago.' },
  { id: 'long-evening', title: 'The Long Evening', desc: 'One last summer evening stretches past midnight, and nobody wants to leave.' },
  { id: 'neon-static', title: 'Neon Static', desc: 'A city street, a missing signal, and a message that only appears in the rain.' },
  { id: 'square-rain', title: 'Square Rain', desc: 'On the wettest night of the year, two strangers miss the same train.' },
  { id: 'after-midnight', title: 'After Midnight', desc: 'The marquee lights stay on long after the last audience has gone home.' },
  { id: 'orbit', title: 'Orbit', desc: 'Ten minutes outside the station. No tether. One chance to look back.' },
  { id: 'docking-day', title: 'Docking Day', desc: 'A routine approach becomes anything but routine when the station goes silent.' },
  { id: 'breakwater', title: 'Breakwater', desc: 'The storm arrives on schedule. The lighthouse keeper does not.' },
  { id: 'whitecap', title: 'Whitecap', desc: 'A boat, a horizon, and a decision that cannot wait for calm water.' },
  { id: 'last-light', title: 'Last Light', desc: 'The sea gives up one last hour of sun before the weather closes in.' },
  { id: 'the-hollow', title: 'The Hollow', desc: 'Something moves between the trees, and the path keeps leading back.' },
  { id: 'ember-road', title: 'Ember Road', desc: 'A road through the fog, a car with no headlights, and a destination nobody names.' },
  { id: 'tall-pines', title: 'Tall Pines', desc: 'The forest is quiet. That is how you know someone is listening.' },
  { id: 'northern-signal', title: 'Northern Signal', desc: 'A green sky over a frozen waterfall, and a signal that answers back.' },
  { id: 'green-sky', title: 'Green Sky', desc: 'The lights come out early tonight, and they are not where they should be.' },
  { id: 'snowbound', title: 'Snowbound', desc: 'A cabin at the edge of the aurora, and a visitor who arrives without footprints.' },
  { id: 'moonlit-runner', title: 'Moonlit Runner', desc: 'A classic car, an open road, and a moon that follows too closely.' },
  { id: 'blue-hour', title: 'Blue Hour', desc: 'Between sunset and dark, a driver makes a deal with the quiet.' },
  { id: 'descent', title: 'Descent', desc: 'Below the surface, the light bends and the rules stop applying.' },
];

const ROUTES = {
  catalogue: { page: 'catalogue', mode: 'default' },
  'catalogue-signed-in': { page: 'catalogue', mode: 'default', signedIn: true },
  'catalogue-loading': { page: 'catalogue', mode: 'loading' },
  'catalogue-empty': { page: 'catalogue', mode: 'empty' },
  'catalogue-poster-error': { page: 'catalogue', mode: 'poster-error' },
  signin: { page: 'signin', target: true, form: 'default' },
  'signin-direct': { page: 'signin', target: false, form: 'default' },
  'signin-required': { page: 'signin', target: true, form: 'required' },
  'signin-error': { page: 'signin', target: true, form: 'error' },
  'signin-loading': { page: 'signin', target: true, form: 'loading' },
  player: { page: 'player', player: 'default', signedIn: true },
  'player-loading': { page: 'player', player: 'loading', signedIn: true },
  'player-error': { page: 'player', player: 'error', signedIn: true },
  'not-found': { page: 'not-found' },
};

const state = {
  page: 'catalogue',
  route: 'catalogue',
  mode: 'default',
  signedIn: false,
  pendingFilm: 'amber-drift',
  playerFilm: 'amber-drift',
  playerState: 'default',
  form: 'default',
  playing: false,
  progress: 0,
  duration: 1440,
  muted: false,
  volume: 70,
};

const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)');

const $ = (id) => document.getElementById(id);
const el = {
  pages: Array.from(document.querySelectorAll('.page')),
  header: $('site-head'),
  navCatalogue: $('nav-catalogue'),
  navBack: $('nav-back'),
  authBtn: $('auth-btn'),
  authLabel: $('auth-label'),
  grid: $('film-grid'),
  gridState: $('catalogue-state'),
  count: $('catalogue-count'),
  lede: document.querySelector('.page-catalogue .lede'),
  chip: document.querySelector('.browse-chip'),
  facts: document.querySelector('.facts-card'),
  factCount: $('fact-count'),
  signinTitle: $('signin-title'),
  authSub: $('auth-sub'),
  targetChip: $('target-chip'),
  targetThumb: $('target-thumb'),
  targetTitle: $('target-title'),
  form: $('signin-form'),
  identifier: $('identifier'),
  password: $('password'),
  fieldId: $('field-id'),
  fieldPass: $('field-pass'),
  formAlert: $('form-alert'),
  submit: $('signin-submit'),
  submitLabel: $('signin-label'),
  cancel: $('cancel-signin'),
  watchTitle: $('watch-title'),
  watchDesc: $('watch-desc'),
  watchIndex: $('watch-index'),
  playerBg: $('player-bg-img'),
  stage: $('stage'),
  stagePoster: $('stage-poster'),
  stagePlay: $('stage-play'),
  stagePlayIcon: $('stage-play-icon'),
  ctlPlay: $('ctl-play'),
  ctlPlayIcon: $('ctl-play-icon'),
  timeCurrent: $('time-current'),
  timeTotal: $('time-total'),
  timeline: $('timeline'),
  mute: $('ctl-mute'),
  muteIcon: $('mute-icon'),
  volume: $('volume'),
  fullscreen: $('ctl-full'),
  retry: $('retry'),
  toast: $('toast'),
  toastText: $('toast-text'),
};

const ICON = {
  play: '<path d="M8 5v14l11-7z"/>',
  pause: '<path d="M7 5h3v14H7zM14 5h3v14h-3z"/>',
  volume:
    '<path d="M4 9v6h4l5 4V5L8 9H4z" fill="currentColor"/><path d="M16.5 8.5a5 5 0 0 1 0 7M19 6a8.5 8.5 0 0 1 0 12" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/>',
  muted:
    '<path d="M4 9v6h4l5 4V5L8 9H4z" fill="currentColor"/><path d="m16 9 5 6M21 9l-5 6" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/>',
  arrow:
    '<svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M7 17 17 7M9 7h8v8" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>',
  film:
    '<svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><rect x="3" y="4" width="18" height="16" rx="2" stroke="currentColor" stroke-width="1.5"/><path d="M7.5 4v16M16.5 4v16M3 9h4.5M3 15h4.5M16.5 9H21M16.5 15H21" stroke="currentColor" stroke-width="1.5"/></svg>',
};

function filmIndex(id) {
  return FILMS.findIndex((f) => f.id === id);
}

function filmById(id) {
  return FILMS.find((f) => f.id === id) || FILMS[0];
}

/* ---------- rendering ---------- */

const LEDE_DEFAULT =
  'Twenty films, streaming instantly. Browse the whole catalogue free — sign in only when you\'re ready to watch.';
const LEDE_EMPTY =
  'The catalogue came back empty. There is nothing to watch right now — check back in a little while.';

function renderCatalogue() {
  const mode = state.mode;
  el.grid.innerHTML = '';
  el.gridState.innerHTML = '';
  el.count.hidden = mode === 'empty';
  el.chip.hidden = mode === 'empty';
  el.facts.hidden = mode === 'empty';
  el.lede.textContent = mode === 'empty' ? LEDE_EMPTY : LEDE_DEFAULT;

  if (mode === 'loading') {
    el.count.textContent = 'Loading';
    el.factCount.textContent = '—';
    el.grid.innerHTML = Array.from({ length: 8 })
      .map(
        (_, i) => `
        <div class="film-card is-skeleton" style="--i:${i}" aria-hidden="true">
          <div class="sk-media"></div>
          <div class="sk-body">
            <div class="sk-line is-title"></div>
            <div class="sk-line"></div>
            <div class="sk-line is-short"></div>
          </div>
        </div>`
      )
      .join('');
    return;
  }

  if (mode === 'empty') {
    el.count.textContent = '0 films';
    el.factCount.textContent = '0';
    el.gridState.innerHTML = `
      <div class="empty-panel">
        <span class="empty-icon">${ICON.film}</span>
        <h3 class="empty-title">No films available</h3>
        <p class="empty-copy">There are no films in the catalogue right now. Please check back soon.</p>
      </div>`;
    return;
  }

  const failures = mode === 'poster-error' ? [3, 7] : [];
  el.count.textContent = `${FILMS.length} films`;
  el.factCount.textContent = String(FILMS.length);
  const verb = state.signedIn ? 'Watch' : 'Sign in to watch';
  el.grid.innerHTML = FILMS.map((film, i) => {
    const media = failures.includes(i)
      ? `<div class="poster-fallback">${ICON.film}<span>Poster unavailable</span></div>`
      : `<img src="assets/img/${film.id}.webp" alt="" decoding="async" />`;
    return `
      <a class="film-card" href="#signin" data-film="${film.id}" style="--i:${i}" aria-label="${verb} ${film.title}">
        <div class="film-media">
          ${media}
          <span class="film-index" aria-hidden="true">${String(i + 1).padStart(2, '0')}</span>
        </div>
        <div class="film-body">
          <h3 class="film-title">${film.title}</h3>
          <p class="film-desc">${film.desc}</p>
          <span class="film-foot"> <span class="arrow-disc">${ICON.arrow}</span></span>
        </div>
      </a>`;
  }).join('');
}

function renderSignin() {
  const film = state.pendingFilm ? filmById(state.pendingFilm) : null;
  const hasTarget = Boolean(film);
  el.targetChip.hidden = !hasTarget;
  el.signinTitle.textContent = hasTarget ? 'Sign in to watch' : 'Sign in to Clover';
  el.authSub.textContent = hasTarget
    ? 'Use the account you were given.'
    : 'Use the account you were given to start watching.';

  if (film) {
    el.targetTitle.textContent = film.title;
    el.targetThumb.src = `assets/img/${film.id}.webp`;
  }

  const loading = state.form === 'loading';
  el.fieldId.classList.toggle('has-error', state.form === 'required' && !el.identifier.value.trim());
  el.fieldPass.classList.toggle('has-error', state.form === 'required' && !el.password.value);
  el.formAlert.classList.toggle('is-shown', state.form === 'error');
  el.identifier.disabled = loading;
  el.password.disabled = loading;
  el.submit.disabled = loading;
  el.submitLabel.innerHTML = loading ? 'Signing in <span class="btn-spinner"></span>' : 'Sign in';
  el.identifier.setAttribute('aria-invalid', String(state.form === 'required' && !el.identifier.value.trim()));
  el.password.setAttribute('aria-invalid', String(state.form === 'required' && !el.password.value));
}

function renderPlayer() {
  const film = filmById(state.playerFilm);
  const index = filmIndex(film.id) + 1;
  el.watchTitle.textContent = film.title;
  el.watchDesc.textContent = film.desc;
  el.watchIndex.textContent = `Film ${String(index).padStart(2, '0')} / ${FILMS.length}`;
  const src = `assets/img/${film.id}.webp`;
  el.playerBg.src = src;
  el.stagePoster.src = src;
  el.stage.classList.toggle('is-loading', state.playerState === 'loading');
  el.stage.classList.toggle('is-error', state.playerState === 'error');
  el.stage.classList.toggle('is-playing', state.playing);
  el.timeTotal.textContent = formatTime(state.duration);
  updatePlaybackUI();
}

function renderHeader() {
  const onCatalogue = state.page === 'catalogue';
  el.navCatalogue.hidden = !onCatalogue;
  el.navBack.hidden = onCatalogue;
  el.navCatalogue.toggleAttribute('aria-current', onCatalogue);
  if (onCatalogue) el.navCatalogue.setAttribute('aria-current', 'page');
  else el.navCatalogue.removeAttribute('aria-current');

  el.authBtn.hidden = state.page === 'signin';
  el.authBtn.className = state.signedIn || !onCatalogue ? 'btn btn-ghost' : 'btn btn-primary';
  el.authLabel.textContent = state.signedIn ? 'Sign out' : 'Sign in';
}

function renderAll() {
  renderHeader();
  renderCatalogue();
  renderSignin();
  renderPlayer();
}

/* ---------- playback ---------- */

let playTimer = null;

function formatTime(seconds) {
  const s = Math.max(0, Math.round(seconds));
  const m = Math.floor(s / 60);
  const rest = s % 60;
  return `${m}:${String(rest).padStart(2, '0')}`;
}

function updatePlaybackUI() {
  const current = (state.progress / 1000) * state.duration;
  el.timeline.value = String(Math.round(state.progress));
  el.timeline.style.setProperty('--p', `${state.progress / 10}%`);
  el.timeCurrent.textContent = formatTime(current);
  el.ctlPlayIcon.innerHTML = state.playing ? ICON.pause : ICON.play;
  el.ctlPlay.setAttribute('aria-label', state.playing ? 'Pause' : 'Play');
  el.stagePlayIcon.innerHTML = state.playing ? ICON.pause : ICON.play;
  el.stagePlay.setAttribute('aria-label', state.playing ? 'Pause' : 'Play');
  el.muteIcon.innerHTML = state.muted ? ICON.muted : ICON.volume;
  el.mute.setAttribute('aria-label', state.muted ? 'Unmute' : 'Mute');
}

function stopTimer() {
  if (playTimer) {
    window.clearInterval(playTimer);
    playTimer = null;
  }
}

function resetPlayback() {
  stopTimer();
  state.playing = false;
  state.progress = 0;
  state.muted = false;
  state.volume = 70;
  if (el.volume) el.volume.value = '70';
}

function setPlaying(on) {
  if (state.playerState !== 'default') return;
  state.playing = on;
  el.stage.classList.toggle('is-playing', on);
  updatePlaybackUI();
  stopTimer();
  if (on) {
    playTimer = window.setInterval(() => {
      state.progress = Math.min(1000, state.progress + 1000 / state.duration);
      if (state.progress >= 1000) {
        setPlaying(false);
        state.progress = 0;
      }
      updatePlaybackUI();
    }, 1000);
  }
}

/* ---------- pages ---------- */

let booted = false;

function arrive(page) {
  page.classList.remove('is-arriving');
  window.clearTimeout(page.arriveTimer);
  void page.offsetWidth;
  if (reduceMotion.matches) return;
  page.classList.add('is-arriving');
  // drop the class once the 520 ms entrance and its last 180 ms stagger are
  // over, so a later re-render of the same page (sign in, sign out, a filter)
  // does not replay it
  page.arriveTimer = window.setTimeout(() => page.classList.remove('is-arriving'), 720);
}

function showPage(pageName) {
  const next = document.querySelector(`[data-page="${pageName}"]`);
  const current = document.querySelector('.page.is-active');
  const first = !booted;
  booted = true;
  if (!next) return;
  if (current === next) {
    next.hidden = false;
    if (first) arrive(next);
    return;
  }
  el.pages.forEach((p) => {
    if (p !== current) {
      p.hidden = true;
      p.classList.remove('is-active', 'is-arriving');
    }
  });
  const reveal = () => {
    next.hidden = false;
    next.classList.add('is-active');
    arrive(next);
    window.scrollTo(0, 0);
  };
  // a cut, not a cross-fade: the reference changes screens in one frame and
  // the new screen's 520 ms entrance does the rest
  if (current) {
    current.hidden = true;
    current.classList.remove('is-active', 'is-arriving');
  }
  reveal();
}

function applyRoute(name) {
  // `/watch/{movieId}` as `#watch-<movieId>`: unknown id is the not-found screen,
  // a known id signed out goes through sign-in and keeps the film.
  const watch = /^watch-(.+)$/.exec(name || '');
  if (watch) {
    const id = decodeURIComponent(watch[1]);
    if (!FILMS.some((f) => f.id === id)) name = 'not-found';
    else if (!state.signedIn) {
      state.pendingFilm = id;
      name = 'signin';
    } else {
      state.playerFilm = id;
      name = 'player';
    }
  }
  const route = ROUTES[name] || ROUTES.catalogue;
  if ('signedIn' in route) state.signedIn = route.signedIn;
  state.route = ROUTES[name] ? name : 'catalogue';
  state.page = route.page;
  state.mode = route.mode || 'default';

  if (route.page === 'signin') {
    if (route.target === false) state.pendingFilm = null;
    else if (!state.pendingFilm) state.pendingFilm = 'amber-drift';
    state.form = route.form || 'default';
    if (state.form === 'error' || state.form === 'loading') {
      if (!el.identifier.value) el.identifier.value = 'viewer@clover.test';
      if (state.form === 'loading') {
        if (!el.password.value) el.password.value = 'clover2026';
      } else {
        el.password.value = '';
      }
    }
    if (state.form === 'default' || state.form === 'required') {
      if (state.form === 'required') {
        el.identifier.value = '';
        el.password.value = '';
      }
    }
  }

  if (route.page === 'player') {
    state.playerState = route.player || 'default';
    resetPlayback();
  }

  showPage(route.page);
  renderAll();
}

/* ---------- toast ---------- */

let toastTimer = null;
function showToast(text) {
  el.toastText.textContent = text;
  el.toast.classList.add('is-shown');
  if (toastTimer) window.clearTimeout(toastTimer);
  toastTimer = window.setTimeout(() => el.toast.classList.remove('is-shown'), 4000);
}

/* ---------- events ---------- */

el.grid.addEventListener('click', (event) => {
  const card = event.target.closest('.film-card[data-film]');
  if (!card) return;
  event.preventDefault();
  const id = card.dataset.film;
  state.pendingFilm = id;
  if (state.signedIn) {
    state.playerFilm = id;
    window.location.hash = '#player';
  } else {
    window.location.hash = '#signin';
  }
});

el.authBtn.addEventListener('click', () => {
  if (state.signedIn) {
    state.signedIn = false;
    showToast('Signed out.');
    if (state.page !== 'catalogue') window.location.hash = '#catalogue';
    else renderAll();
  } else {
    window.location.hash = '#signin-direct';
  }
});

el.form.addEventListener('submit', (event) => {
  event.preventDefault();
  const id = el.identifier.value.trim();
  const pw = el.password.value;
  if (!id || !pw) {
    state.form = 'required';
    renderSignin();
    (id ? el.password : el.identifier).focus();
    return;
  }
  state.form = 'loading';
  renderSignin();

  window.setTimeout(() => {
    if (pw === 'wrongpass') {
      state.form = 'error';
      el.password.value = '';
      renderSignin();
      el.password.focus();
      return;
    }
    const target = state.pendingFilm;
    state.signedIn = true;
    state.form = 'default';
    if (target) {
      state.playerFilm = target;
      showToast('Signed in.');
      window.location.hash = '#player';
    } else {
      showToast('Signed in. Pick a film to watch.');
      window.location.hash = '#catalogue';
    }
  }, 700);
});

el.cancel.addEventListener('click', () => {
  window.location.hash = '#catalogue';
});

el.stagePlay.addEventListener('click', () => setPlaying(!state.playing));
el.ctlPlay.addEventListener('click', () => setPlaying(!state.playing));

el.timeline.addEventListener('input', () => {
  state.progress = Number(el.timeline.value);
  updatePlaybackUI();
});

el.volume.addEventListener('input', () => {
  state.volume = Number(el.volume.value);
  state.muted = state.volume === 0;
  updatePlaybackUI();
});

el.mute.addEventListener('click', () => {
  state.muted = !state.muted;
  el.volume.value = state.muted ? '0' : String(state.volume || 70);
  updatePlaybackUI();
});

el.fullscreen.addEventListener('click', async () => {
  try {
    if (!document.fullscreenElement) await el.stage.requestFullscreen();
    else await document.exitFullscreen();
  } catch {
    /* the mock keeps working when fullscreen is blocked */
  }
});

document.addEventListener('fullscreenchange', () => {
  el.fullscreen.setAttribute(
    'aria-label',
    document.fullscreenElement ? 'Exit fullscreen' : 'Enter fullscreen'
  );
});

el.retry.addEventListener('click', () => {
  state.playerState = 'loading';
  renderPlayer();
  window.history.replaceState(null, '', '#player-loading');
  window.setTimeout(() => {
    state.playerState = 'default';
    window.history.replaceState(null, '', '#player');
    renderPlayer();
    showToast('Film ready.');
  }, 1400);
});

document.querySelectorAll('[data-scroll]').forEach((button) => {
  button.addEventListener('click', () => {
    const target = document.getElementById(button.dataset.scroll);
    if (target) target.scrollIntoView({ behavior: reduceMotion.matches ? 'auto' : 'smooth' });
  });
});

window.addEventListener('scroll', () => {
  el.header.classList.toggle('is-solid', window.scrollY > 24);
});

window.addEventListener('hashchange', () => {
  applyRoute(window.location.hash.slice(1) || 'catalogue');
});

/* initial screen */
applyRoute(window.location.hash.slice(1) || 'catalogue');
