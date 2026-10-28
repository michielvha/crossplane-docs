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
const sourceProgram = '{{ range .spec.networks }}\nkind: ConfigMap\n# literal \\n stays literal\n{{ end }}\n</script><img src=x onerror=alert(1)>\n';
function generateHTML(layout: string) {
 const result = Bun.spawnSync([binary, 'generate', bundle, '--all-versions', '--format', 'html', '--html-layout', layout]);
 expect(result.exitCode).toBe(0);
 return result.stdout.toString();
}
function generate(format: string) {
 const result = Bun.spawnSync([binary, 'generate', join(root, 'examples/platform.yaml'), '--all-versions', '--format', format]);
 expect(result.exitCode).toBe(0);
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
 return result.stdout.toString();
}
beforeAll(async () => {
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
 expect(await page.locator('h1').textContent()).toContain('Crossplane');
 expect(await page.locator('table').count()).toBeGreaterThan(0);
 expect(await page.locator('th[scope="col"]').count()).toBeGreaterThan(0);
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
   await page.locator('.api-panel:visible .field-link[data-path="spec.region"]').click();
   expect(await page.locator('.api-panel:visible .field-entry:visible').textContent()).toContain('spec.region');
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
