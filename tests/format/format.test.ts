import { test, expect, beforeAll, afterAll } from 'bun:test';
import { convert } from '@asciidoctor/core';
import { chromium, type Browser } from 'playwright';
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';

const root = resolve(import.meta.dir, '../..');
const work = mkdtempSync(join(tmpdir(), 'crossplane-docs-format-'));
const binary = join(work, 'crossplane-docs');
let browser: Browser;
const bundle = join(work, 'bundle.yaml');
const sourceProgram = '{{ range .spec.networks }}\nkind: ConfigMap\r\n\t# Unicode λ; literal \\n stays literal\n{{ end }}\n</script><img src=x onerror=alert(1)>\n';
function generateHTML(layout: string) {
 const result = Bun.spawnSync([binary, 'generate', bundle, '--all-versions', '--format', 'html', '--html-layout', layout]);
 expect(result.exitCode).toBe(0);
 return result.stdout.toString();
}
function generate(format: string) {
 const result = Bun.spawnSync([binary, 'generate', join(root, 'examples/platform.yaml'), '--all-versions', '--format', format]);
 expect(result.exitCode).toBe(0);
 return result.stdout.toString();
}
function writeBundle() {
 const files = ['examples/platform.yaml', 'examples/platform/go-templating.yaml', 'examples/platform/xr.yaml'];
 writeFileSync(bundle, files.map(file => readFileSync(join(root, file), 'utf8')).join('\n---\n') + `
---
apiVersion: apiextensions.crossplane.io/v2
kind: CompositeResourceDefinition
metadata: {name: xplatforms.other.example.org}
spec:
  group: other.example.org
  names: {kind: XPlatform, plural: xplatforms}
  scope: Namespaced
  versions:
  - name: v1beta1
    served: true
    referenceable: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              owner: {type: string, description: Other API owner}
---
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata: {name: source-readability}
spec:
  compositeTypeRef: {apiVersion: platform.example.org/v1beta1, kind: XPlatform}
  mode: Pipeline
  pipeline:
  - step: multiline
    functionRef: {name: custom}
    input:
      script: ${JSON.stringify(sourceProgram)}
`);
}
beforeAll(async () => {
 writeBundle();
 const result = Bun.spawnSync(['go', 'build', '-o', binary, '.'], {cwd: root});
 expect(result.exitCode).toBe(0);
 browser = await chromium.launch({headless: true});
}, 120000);
afterAll(async () => {await browser?.close(); rmSync(work, {recursive: true, force: true});});

test('standalone HTML works offline on desktop and mobile', async () => {
 const html = generate('html');
 const page = await browser.newPage();
 await page.route(/^https?:/, route => route.abort());
 const errors: string[] = [];
 page.on('pageerror', error => errors.push(error.message));
 await page.setContent(html);
 expect(await page.locator('h1:visible').textContent()).toContain('XPlatform');
 expect(await page.locator('.field-entry').count()).toBeGreaterThan(0);
 expect(await page.locator('label[for="field-search"]').count()).toBe(1);
 expect(await page.locator('script[src], link[rel="stylesheet"]').count()).toBe(0);
 for (const width of [1280, 390]) {
  await page.setViewportSize({width, height: 800});
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
 }
 const broken = await page.locator('nav a').evaluateAll(links => links.filter(a => !document.querySelector(a.getAttribute('href')!)).length);
 expect(broken).toBe(0);
 expect(errors).toEqual([]);
 await page.close();
});

for (const layout of ['reference', 'explorer', 'compact']) {
 test(`${layout}: readable offline layout, navigation and mobile interactions`, async () => {
  const output = generateHTML(layout);
  const page = await browser.newPage({viewport: {width: 1280, height: 900}});
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await page.route(/^https?:/, route => route.abort());
  await page.setContent(output);
  expect(await page.locator('body').getAttribute('data-html-layout')).toBe(layout);
  expect(await page.locator('.metrics').count()).toBe(0);
  expect(await page.locator('script[src],link[rel="stylesheet"]').count()).toBe(0);
  expect(await page.locator('code.source-code').allTextContents()).toContain(sourceProgram);
  expect(await page.locator('img[src="x"]').count()).toBe(0);
  const duplicates = await page.locator('[id]').evaluateAll(nodes => {
   const ids = nodes.map(n => n.id); return ids.filter((id, i) => ids.indexOf(id) !== i);
  });
  expect(duplicates).toEqual([]);
  await page.evaluate(() => {
   Object.defineProperty(navigator, 'clipboard', {configurable: true, value: {writeText: async (value: string) => { (window as any).copiedSource = value; }}});
  });
  const programBlock = page.locator('.code-block').filter({has: page.locator('code.source-code')}).last();
  await programBlock.locator('[data-copy]').click();
  expect(await page.evaluate(() => (window as any).copiedSource)).toBe(sourceProgram);
  await page.locator('#field-search').fill('region');
  expect(await page.locator('#search-status').textContent()).toContain('matching');
  await page.locator('#field-search').fill('does-not-exist');
  expect(await page.locator('.empty-fields:visible').count()).toBe(1);
  await page.locator('#field-search').fill('');
  await page.locator('#api-select').selectOption({label: 'XPlatform / other.example.org'});
  expect(await page.locator('.api-panel:visible').textContent()).toContain('Other API owner');
  await page.locator('#api-select').selectOption({label: 'XPlatform / platform.example.org'});
  await page.locator('#version-select').selectOption({label: 'v1alpha1'});
  expect(await page.locator('.api-panel:visible').getAttribute('data-version')).toBe('v1alpha1');
  await page.locator('#version-select').selectOption({label: 'v1beta1'});
  if (layout === 'explorer') {
   await page.locator('.api-panel:visible .field-link[data-path="spec.storage.sizeGiB"]').click();
   expect(await page.locator('.api-panel:visible .field-entry:visible').textContent()).toContain('sizeGiB');
  }
  for (const width of [1280, 768, 390, 360]) {
   await page.setViewportSize({width, height: 900});
   expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
   const overflowing = await page.locator('.field-entry:visible').evaluateAll(nodes => nodes.filter(n => n.scrollWidth > n.clientWidth + 1).length);
   expect(overflowing).toBe(0);
  }
  if (layout === 'reference') {
   await page.locator('.api-panel:visible [data-view="example"]').click();
   expect(await page.locator('.api-panel:visible .paired-example').isVisible()).toBe(true);
   await page.locator('.api-panel:visible [data-view="fields"]').click();
   expect(await page.locator('.api-panel:visible .field-list').isVisible()).toBe(true);
  }
  if (layout === 'explorer') {
   await page.locator('.api-panel:visible [data-browse-fields]').click();
   expect(await page.locator('.api-panel:visible .field-browser').isVisible()).toBe(true);
   await page.locator('#field-search').fill('no-matching-field');
   expect(await page.locator('.api-panel:visible .empty-fields').isVisible()).toBe(true);
   await page.locator('#field-search').fill('');
   await page.locator('.api-panel:visible .field-link[data-path="spec.region"]').click();
   expect(await page.locator('.api-panel:visible .field-entry:visible').textContent()).toContain('spec.region');
   expect(await page.locator('.api-panel:visible [data-browse-fields]').textContent()).toBe('Browse fields');
   expect(await page.locator('.api-panel:visible [data-browse-fields]').getAttribute('aria-expanded')).toBe('false');
  }
  expect(errors).toEqual([]);
  await page.close();
  const plain = await browser.newPage({javaScriptEnabled: false, viewport: {width: 390, height: 900}});
  await plain.setContent(output);
  expect(await plain.locator('.api-panel:visible').count()).toBe(3);
  expect(await plain.getByText('Other API owner', {exact: true}).count()).toBeGreaterThan(0);
  expect(await plain.locator('code.source-code').allTextContents()).toContain(sourceProgram);
  await plain.close();
 });
}

test('field links, search and API selection survive reload without a stale hash', async () => {
 const output = generateHTML('explorer');
 const page = await browser.newPage();
 await page.route('https://reference.test/**', route => route.fulfill({body: output, contentType: 'text/html'}));
 await page.goto('https://reference.test/');
 await page.locator('#api-select').selectOption({label: 'XPlatform / platform.example.org'});
 const field = page.locator('.api-panel:visible .field-link[data-path="spec.storage.sizeGiB"]');
 const target = await field.getAttribute('href');
 await page.goto('https://reference.test/' + target);
 expect(await page.locator('.api-panel:visible .field-entry:visible').getAttribute('data-path')).toBe('spec.storage.sizeGiB');
 await page.locator('#version-select').selectOption({label: 'v1alpha1'});
 await page.reload();
 expect(await page.locator('.api-panel:visible').getAttribute('data-version')).toBe('v1alpha1');
 await page.locator('#field-search').fill('region');
 await page.reload();
 expect(await page.locator('#field-search').inputValue()).toBe('region');
 expect(await page.locator('.api-panel:visible .field-entry:visible').getAttribute('data-path')).toContain('region');
 await page.close();
});

test('mobile explorer reports empty searches and synchronizes its browse action', async () => {
 const page = await browser.newPage({viewport: {width: 390, height: 844}});
 await page.setContent(generateHTML('explorer'));
 await page.locator('#api-select').selectOption({label: 'XPlatform / platform.example.org'});
 const browse = page.locator('.api-panel:visible [data-browse-fields]');
 await browse.click();
 await page.locator('#field-search').fill('no-matching-field');
 expect(await page.locator('.api-panel:visible .empty-fields').isVisible()).toBe(true);
 await page.locator('#field-search').fill('');
 await page.locator('.api-panel:visible .field-link[data-path="spec.region"]').click();
 expect(await browse.textContent()).toBe('Browse fields');
 expect(await browse.getAttribute('aria-expanded')).toBe('false');
 await browse.click();
 await page.locator('#version-select').selectOption({label: 'v1alpha1'});
 expect(await page.locator('.api-panel:visible [data-browse-fields]').textContent()).toBe('Browse fields');
 await page.close();
});

test('AsciiDoc is consumable by Asciidoctor', async () => {
 const output = (await convert(generate('asciidoc'), {standalone: true, safe: 'secure'})).toString();
 expect(output).toContain('<table');
 expect(output).toContain('spec.networks[].cidr');
 expect(output).toContain('XPlatform');
});

test('custom HTML templates escape manifest content', () => {
 const template = join(work, 'custom.html');
 writeFileSync(template, '<h1>{{ .Title }}</h1>{{range .APIs}}<p>{{.Description}}</p>{{end}}');
 const config = join(work, 'config.yml');
 writeFileSync(config, 'version: 1\nformatter: html\ntitle: "<script>bad</script>"\n');
 const result = Bun.spawnSync([binary, 'generate', join(root, 'examples/platform.yaml'), '--config', config, '--template', template]);
 expect(result.exitCode).toBe(0);
 expect(result.stdout.toString()).toContain('&lt;script&gt;');
 expect(result.stdout.toString()).not.toContain('<script>');
});

for (const layout of ['reference', 'explorer', 'compact']) {
 test(`${layout}: requirement filters combine with search and preserve navigation state`, async () => {
  const config = join(work, `filter-${layout}.yml`);
  writeFileSync(config, `version: 1\nformatter: html\nhtml-layout: ${layout}\ncolumns: [path, type, description]\n`);
  const result = Bun.spawnSync([binary, 'generate', bundle, '--all-versions', '--config', config]);
  expect(result.exitCode).toBe(0);
  const page = await browser.newPage();
  await page.route('https://reading.test/**', route => route.fulfill({body: result.stdout.toString(), contentType: 'text/html'}));
  await page.goto('https://reading.test/');
  await page.locator('#api-select').selectOption({label: 'XPlatform / platform.example.org'});
  const requirement = page.getByLabel('Field requirement', {exact: true});
  expect(await requirement.count()).toBe(1);
  await requirement.selectOption('required');
  const active = page.locator('.api-panel:visible');
  const shown = layout === 'explorer' ? active.locator('.field-link:not([hidden])') : active.locator('.field-entry:not([hidden])');
  expect(await shown.evaluateAll(nodes => nodes.map(n => (n as HTMLElement).dataset.path).sort())).toEqual(['spec.networks', 'spec.networks[].cidr', 'spec.networks[].name', 'spec.region', 'spec.variant.oneOf[0].managed', 'spec.variant.oneOf[1].externalID']);
  expect(await page.locator('#search-status').textContent()).toBe('6 matching fields');
  await page.locator('#field-search').fill('region');
  expect(await shown.count()).toBe(1);
  await requirement.selectOption('optional');
  expect(await shown.count()).toBe(1); // status.observedRegion is optional.
  await page.locator('#field-search').fill('optionalNote');
  expect(await shown.count()).toBe(1);
  await page.reload();
  expect(await requirement.inputValue()).toBe('optional');
  expect(await page.locator('#field-search').inputValue()).toBe('optionalNote');
  await page.locator('#version-select').selectOption({label: 'v1alpha1'});
  expect(await requirement.inputValue()).toBe('optional');
  expect(await shown.count()).toBe(1);
  await page.locator('#api-select').selectOption({label: 'XPlatform / other.example.org'});
  expect(await active.locator('.empty-fields').isVisible()).toBe(true);
  expect(await page.locator('#search-status').textContent()).toBe('0 matching fields');
  await page.locator('#api-select').selectOption({label: 'XPlatform / platform.example.org'});
  const target = await active.locator('.field-entry[data-path="spec.region"]').getAttribute('id');
  await page.goto(`https://reading.test/?requirement=optional&q=optionalNote#${target}`);
  expect(await requirement.inputValue()).toBe('all');
  expect(await page.locator('#field-search').inputValue()).toBe('');
  expect(await active.locator('.field-entry[data-path="spec.region"]').isVisible()).toBe(true);
  if (layout === 'explorer') {
   const api = await active.getAttribute('id');
   for (const filters of ['requirement=optional', 'q=optionalNote', 'requirement=optional&q=optionalNote']) {
    await page.goto(`https://reading.test/?api=${api}&field=spec.region&${filters}`);
    expect(await active.locator('.field-entry:visible').getAttribute('data-path')).toBe('spec.region');
    expect(await requirement.inputValue()).toBe('all');
    expect(await page.locator('#field-search').inputValue()).toBe('');
    expect(new URL(page.url()).searchParams.get('field')).toBe('spec.region');
   }
  }
  await page.goto('https://reading.test/?requirement=unknown');
  expect(await requirement.inputValue()).toBe('all');
  expect(new URL(page.url()).searchParams.has('requirement')).toBe(false);
  await page.close();
 });
}

test('theme follows the system until a keyboard choice, remembers it, and tolerates denied storage', async () => {
 const context = await browser.newContext({colorScheme: 'dark'});
 const page = await context.newPage();
 await page.route('https://reading.test/**', route => route.fulfill({body: generateHTML('reference'), contentType: 'text/html'}));
 await page.goto('https://reading.test/');
 const toggle = page.getByRole('button', {name: 'Dark mode', exact: true});
 expect(await toggle.count()).toBe(1);
 expect(await toggle.getAttribute('aria-pressed')).toBe('true');
 await page.emulateMedia({colorScheme: 'light'});
 await page.waitForFunction(() => document.documentElement.dataset.theme === 'light');
 expect(await toggle.getAttribute('aria-pressed')).toBe('false');
 await toggle.focus();
 await page.keyboard.press('Enter');
 expect(await toggle.getAttribute('aria-pressed')).toBe('true');
 await page.reload();
 expect(await toggle.getAttribute('aria-pressed')).toBe('true');
 await page.emulateMedia({colorScheme: 'dark'});
 await toggle.click();
 await page.reload();
 expect(await toggle.getAttribute('aria-pressed')).toBe('false');
 await page.close();
 const denied = await context.newPage();
 await denied.addInitScript(() => { Object.defineProperty(window, 'localStorage', {get() {throw new DOMException('Denied', 'SecurityError');}}); });
 const errors: string[] = [];
 denied.on('pageerror', e => errors.push(e.message));
 await denied.setContent(generateHTML('compact'));
 await denied.getByRole('button', {name: 'Dark mode', exact: true}).click();
 expect(errors).toEqual([]);
 expect(await denied.locator('#field-search').isVisible()).toBe(true);
 await denied.close();
 const composition = Bun.spawnSync([binary, 'composition', join(root, 'examples/platform/go-templating.yaml'), '--format', 'html']);
 expect(composition.exitCode).toBe(0);
 const file = join(work, 'composition.html');
 writeFileSync(file, composition.stdout);
 const local = await context.newPage();
 await local.goto('file://' + file);
 expect(await local.locator('.api-panel').count()).toBe(0);
 const localToggle = local.getByRole('button', {name: 'Dark mode', exact: true});
 expect(await localToggle.isVisible()).toBe(true);
 await localToggle.click();
 expect(await localToggle.getAttribute('aria-pressed')).toBe('false');
 await local.close();
 await context.close();
});

function contrast(a: string, b: string) {
 const luminance = (color: string) => {
  const parts = color.match(/[\d.]+/g)!.slice(0, 3).map(Number).map(v => { const x = v / 255; return x <= .04045 ? x / 12.92 : ((x + .055) / 1.055) ** 2.4; });
  return parts[0] * .2126 + parts[1] * .7152 + parts[2] * .0722;
 };
 const x = luminance(a), y = luminance(b);
 return (Math.max(x, y) + .05) / (Math.min(x, y) + .05);
}

for (const layout of ['reference', 'explorer', 'compact']) {
 test(`${layout}: midnight colours, contrast, native controls and mobile geometry`, async () => {
  const page = await browser.newPage({colorScheme: 'light'});
  await page.setContent(generateHTML(layout));
  await page.locator('#api-select').selectOption({label: 'XPlatform / platform.example.org'});
  const toggle = page.getByRole('button', {name: 'Dark mode', exact: true});
  expect(await toggle.count()).toBe(1);
  const palette = async () => page.evaluate(() => {
   const style = (selector: string) => getComputedStyle(document.querySelector(selector)!);
   const focus = document.querySelector('#theme-toggle') as HTMLElement;
   return {ink: style('body').color, background: style('body').backgroundImage, scheme: style('html').colorScheme,
    control: style('#field-search').backgroundColor, controlInk: style('#field-search').color, border: style('#field-search').borderTopColor,
    muted: style('.document-title').color, code: style('.code-block pre').color, accent: style('.brand span').color,
    outline: getComputedStyle(focus).outlineColor, meta: document.querySelector('meta[name="theme-color"]')?.getAttribute('content')};
  });
  expect((await palette()).ink).toBe('rgb(23, 42, 57)');
  await toggle.click();
  await toggle.focus();
  const colors = await palette();
  expect(colors.background).toBe('linear-gradient(160deg, rgb(2, 8, 20) 0%, rgb(6, 18, 41) 48%, rgb(11, 31, 69) 100%)');
  expect(colors.scheme).toBe('dark');
  expect(colors.meta).toBe('#020814');
  // Brightest gradient stop and two nested translucent white surfaces.
  const surface = 'rgb(32, 51, 85)';
  for (const color of [colors.ink, colors.muted, colors.code, colors.accent]) expect(contrast(color, surface)).toBeGreaterThanOrEqual(4.5);
  expect(contrast(colors.controlInk, colors.control)).toBeGreaterThanOrEqual(4.5);
  expect(contrast(colors.border, colors.control)).toBeGreaterThanOrEqual(3);
  expect(contrast(colors.outline, colors.control)).toBeGreaterThanOrEqual(3);
  for (const width of [360, 390, 768, 1280]) {
   await page.setViewportSize({width, height: 900});
   expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
   expect(await toggle.evaluate(el => el.getBoundingClientRect().height >= 44)).toBe(true);
   await page.getByLabel('Field requirement', {exact: true}).selectOption('required');
   expect(await page.locator('.api-panel:visible .empty-fields').isVisible()).toBe(false);
  }
  await toggle.click();
  expect((await palette()).ink).toBe('rgb(23, 42, 57)');
  await page.close();
 });
}

test('no JavaScript uses system colours and print reveals all filtered fields in light', async () => {
 for (const layout of ['reference', 'explorer', 'compact']) {
  for (const colorScheme of ['light', 'dark'] as const) {
   const page = await browser.newPage({javaScriptEnabled: false, colorScheme});
   await page.setContent(generateHTML(layout));
   expect(await page.getByRole('button', {name: 'Dark mode', exact: true}).count()).toBe(0);
   expect(await page.locator('.controls').isVisible()).toBe(false);
   expect(await page.locator('.api-panel:visible').count()).toBe(3);
   expect(await page.locator('body').evaluate(el => getComputedStyle(el).color)).toBe(colorScheme === 'dark' ? 'rgb(234, 240, 255)' : 'rgb(23, 42, 57)');
   const effectiveMeta = await page.locator('meta[name="theme-color"]').evaluateAll(nodes => nodes.filter(node => !node.getAttribute('media') || matchMedia(node.getAttribute('media')!).matches).map(node => node.getAttribute('content')));
   expect(effectiveMeta).toEqual([colorScheme === 'dark' ? '#020814' : '#f5f8f8']);
   await page.close();
  }
  const page = await browser.newPage({colorScheme: 'dark'});
  await page.setContent(generateHTML(layout));
  expect(await page.getByLabel('Field requirement', {exact: true}).count()).toBe(1);
  await page.getByLabel('Field requirement', {exact: true}).selectOption('required');
  await page.locator('#field-search').fill('nonexistent');
  await page.emulateMedia({media: 'print'});
  expect(await page.locator('body').evaluate(el => getComputedStyle(el).color)).toBe('rgb(23, 42, 57)');
  expect(await page.locator('.api-panel:visible').count()).toBe(3);
  expect(await page.locator('.field-entry:visible').count()).toBe(await page.locator('.field-entry').count());
  expect(await page.locator('.empty-fields:visible').count()).toBe(0);
  await page.close();
 }
});
