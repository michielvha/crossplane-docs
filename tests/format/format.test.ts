import { test, expect, beforeAll, afterAll } from 'bun:test';
import { convert } from '@asciidoctor/core';
import { chromium, type Browser } from 'playwright';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';

const root = resolve(import.meta.dir, '../..');
const work = mkdtempSync(join(tmpdir(), 'crossplane-docs-format-'));
const binary = join(work, 'crossplane-docs');
let browser: Browser;
function generate(format: string) {
 const result = Bun.spawnSync([binary, 'generate', join(root, 'examples/platform.yaml'), '--all-versions', '--format', format]);
 expect(result.exitCode).toBe(0);
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
