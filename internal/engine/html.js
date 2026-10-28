// All facts are server-rendered. Enhancement only changes which content is in view.
const panels = [...document.querySelectorAll('.api-panel')];
const controls = document.querySelector('.controls');
const resourceSelect = document.getElementById('api-select');
const versionSelect = document.getElementById('version-select');
const versions = versionSelect ? [...versionSelect.options] : [];
const search = document.getElementById('field-search');
const requirement = document.getElementById('requirement-filter');
const status = document.getElementById('search-status');
const layout = document.body.dataset.htmlLayout;
const params = new URLSearchParams(location.search);
let panel;
let selectedPath = params.get('field') || '';

function syncBrowseAction() {
  const button = panel?.querySelector('[data-browse-fields]');
  if (!button) return;
  const browsing = panel.classList.contains('browsing');
  button.textContent = browsing ? 'Return to field' : 'Browse fields';
  button.setAttribute('aria-expanded', String(browsing));
}

function clearFieldHash() {
  if (!['http:', 'https:', 'file:'].includes(location.protocol)) return;
  const url = new URL(location.href);
  url.hash = '';
  try { history.replaceState(null, '', url); } catch { /* Embedded preview. */ }
}

function saveState() {
  if (!panel || !['http:', 'https:', 'file:'].includes(location.protocol)) return;
  const url = new URL(location.href);
  url.searchParams.set('api', panel.id);
  if (search.value) url.searchParams.set('q', search.value); else url.searchParams.delete('q');
  if (requirement.value !== 'all') url.searchParams.set('requirement', requirement.value); else url.searchParams.delete('requirement');
  if (layout === 'explorer' && selectedPath) url.searchParams.set('field', selectedPath); else url.searchParams.delete('field');
  try { history.replaceState(null, '', url); } catch { /* Embedded previews may prohibit URL changes. */ }
}

function filterFields() {
  if (!panel) return;
  const query = search.value.trim().toLowerCase();
  const entries = [...panel.querySelectorAll('.field-entry')];
  const matches = entries.filter(entry => entry.dataset.search.includes(query) && (requirement.value === 'all' || entry.dataset.required === String(requirement.value === 'required')));
  const links = [...panel.querySelectorAll('.field-link')];
  let selected = matches.find(entry => entry.dataset.path === selectedPath);
  if (!selected) selected = matches.find(entry => entry.dataset.path === 'spec.region') || matches[0];
  if (selected) selectedPath = selected.dataset.path;
  for (const entry of entries) entry.hidden = !matches.includes(entry) || (layout === 'explorer' && entry !== selected);
  for (const group of panel.querySelectorAll('.field-group')) {
    group.hidden = !group.querySelector('.field-entry:not([hidden])');
    if ((query || requirement.value !== 'all') && !group.hidden) group.open = true;
  }
  for (const link of links) {
    link.hidden = !matches.some(entry => entry.id === link.hash.slice(1));
    link.setAttribute('aria-current', String(!!selected && link.hash.slice(1) === selected.id));
  }
  const empty = panel.querySelector('.empty-fields');
  if (empty) empty.hidden = matches.length > 0;
  status.textContent = query || requirement.value !== 'all' ? `${matches.length} matching fields` : '';
  saveState();
}

function selectPanel(id) {
  panel = panels.find(item => item.id === id) || panels.find(item => item.dataset.preferred === 'true') || panels[0];
  if (!panel) return;
  for (const item of panels) item.hidden = item !== panel;
  const focused = document.activeElement;
  panel.querySelector('.api-header').insertAdjacentElement('afterend', controls);
  if (focused && controls.contains(focused)) focused.focus({preventScroll: true});
  resourceSelect.value = panel.dataset.resource;
  versionSelect.replaceChildren(...versions.filter(option => option.dataset.resource === panel.dataset.resource).map(option => option.cloneNode(true)));
  versionSelect.value = panel.id;
  panel.classList.remove('browsing', 'example-view');
  syncBrowseAction();
  for (const button of panel.querySelectorAll('[data-view]')) button.setAttribute('aria-pressed', String(button.dataset.view === 'fields'));
  for (const link of document.querySelectorAll('.api-nav')) link.setAttribute('aria-current', String(link.hash === '#' + panel.id));
  filterFields();
}

function revealHash() {
  let id;
  try { id = decodeURIComponent(location.hash.slice(1)); } catch { return; }
  const target = document.getElementById(id);
  const owner = target?.closest('.api-panel');
  if (!owner) return;
  if (target.classList.contains('field-entry')) {
    selectedPath = target.dataset.path;
    search.value = '';
    requirement.value = 'all';
  }
  selectPanel(owner.id);
  for (let parent = target.parentElement; parent; parent = parent.parentElement) if (parent.tagName === 'DETAILS') parent.open = true;
  target.scrollIntoView({block: 'start'});
}

if (panels.length && controls) {
  document.documentElement.classList.add('js');
  controls.hidden = false;
  const resources = new Set(panels.map(item => item.dataset.resource));
  if (resources.size === 1) resourceSelect.closest('.control').hidden = true;
  search.value = params.get('q') || '';
  requirement.value = ['required', 'optional'].includes(params.get('requirement')) ? params.get('requirement') : 'all';
  selectPanel(params.get('api'));
  revealHash();
  search.addEventListener('input', () => { clearFieldHash(); filterFields(); });
  requirement.addEventListener('change', () => { clearFieldHash(); filterFields(); });
  versionSelect.addEventListener('change', () => { clearFieldHash(); selectPanel(versionSelect.value); });
  resourceSelect.addEventListener('change', () => {
    clearFieldHash();
    const candidates = panels.filter(item => item.dataset.resource === resourceSelect.value);
    selectPanel((candidates.find(item => item.dataset.preferred === 'true') || candidates[0]).id);
  });
  for (const link of document.querySelectorAll('.api-nav')) link.addEventListener('click', () => { selectPanel(link.hash.slice(1)); });
  for (const link of document.querySelectorAll('.field-link')) link.addEventListener('click', event => {
    event.preventDefault();
    clearFieldHash();
    selectedPath = link.dataset.path;
    panel.classList.remove('browsing');
    syncBrowseAction();
    filterFields();
    const target = document.getElementById(link.hash.slice(1));
    const heading = target?.querySelector('h3');
    if (heading) heading.focus({preventScroll: true});
    if (innerWidth <= 680) target?.scrollIntoView({block: 'start'});
  });
  for (const button of document.querySelectorAll('[data-browse-fields]')) {
    button.hidden = false;
    button.addEventListener('click', () => {
      const browsing = panel.classList.toggle('browsing');
      syncBrowseAction();
      if (browsing) panel.querySelector('.field-link:not([hidden])')?.focus({preventScroll: true});
    });
  }
  for (const views of document.querySelectorAll('.mobile-views')) views.hidden = false;
  for (const button of document.querySelectorAll('[data-view]')) button.addEventListener('click', () => {
    panel.classList.toggle('example-view', button.dataset.view === 'example');
    for (const sibling of panel.querySelectorAll('[data-view]')) sibling.setAttribute('aria-pressed', String(sibling === button));
  });
  addEventListener('hashchange', revealHash);
}

const themeToggle = document.getElementById('theme-toggle');
if (themeToggle) {
  const root = document.documentElement;
  const systemTheme = matchMedia('(prefers-color-scheme: dark)');
  function applyTheme(theme) {
    root.dataset.theme = theme;
    themeToggle.setAttribute('aria-pressed', String(theme === 'dark'));
    themeToggle.title = theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode';
    document.querySelector('meta[name="theme-color"]').content = theme === 'dark' ? '#020814' : '#f5f8f8';
  }
  applyTheme(root.dataset.theme);
  themeToggle.hidden = false;
  systemTheme.addEventListener('change', event => {
    if (root.dataset.themePreference === 'system') applyTheme(event.matches ? 'dark' : 'light');
  });
  themeToggle.addEventListener('click', () => {
    const theme = root.dataset.theme === 'dark' ? 'light' : 'dark';
    root.dataset.themePreference = theme;
    applyTheme(theme);
    try { localStorage.setItem('crossplane-docs-theme', theme); } catch { /* The choice still applies for this document. */ }
  });
}

let printClosedDetails = [];
addEventListener('beforeprint', () => {
  printClosedDetails = [...document.querySelectorAll('details:not([open])')];
  for (const details of printClosedDetails) details.open = true;
});
addEventListener('afterprint', () => {
  for (const details of printClosedDetails) details.open = false;
  printClosedDetails = [];
});

for (const button of document.querySelectorAll('.code-block [data-copy]')) {
  button.hidden = false;
  button.addEventListener('click', async () => {
    const block = button.closest('.code-block');
    const code = block.querySelector('code');
    const feedback = block.querySelector('.copy-status');
    try {
      await navigator.clipboard.writeText(code.textContent);
      feedback.textContent = 'Copied.';
    } catch {
      const selection = getSelection();
      const range = document.createRange();
      range.selectNodeContents(code);
      selection.removeAllRanges();
      selection.addRange(range);
      feedback.textContent = 'Select the highlighted code and copy it.';
    }
  });
}
