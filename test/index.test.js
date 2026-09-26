// Runs against a real build in dist/, and the compiled loader in lib/.
import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import { before, test } from "node:test";
import { fileURLToPath } from "node:url";

import { loadImages } from "../lib/index.js";

const RIFTER = 587;
const RIFTER_BLUEPRINT = 691;
const DAMAGE_CONTROL_II = 2048;
const DAMAGE_CONTROL_II_BLUEPRINT = 2049;
const INTACT_HULL_SECTION = 30752;
const TECH_II = 2;
const SHIPS = 4;

let images;

before(async () => {
  images = await loadImages(
    { bytes: readFileSync(new URL("../dist/images.dat", import.meta.url)) },
    { baseUrl: new URL("../dist/images", import.meta.url) },
  );
});

function assertExists(src) {
  assert.ok(existsSync(fileURLToPath(src)), `${src} does not exist`);
}

test("a ship has one layer", () => {
  const layers = images.typeIcon(RIFTER);
  assert.equal(layers.length, 1);
  assert.deepEqual(Object.keys(layers[0]), ["src"]);
  assertExists(layers[0].src);
});

test("a tech II module has a marker, which can be left off", () => {
  const layers = images.typeIcon(DAMAGE_CONTROL_II);
  assert.equal(layers.length, 2);
  assert.equal(layers[1].src, images.metaGroupIcon(TECH_II));

  assert.deepEqual(images.typeIcon(DAMAGE_CONTROL_II, { marker: false }), layers.slice(0, 1));
});

test("a ship blueprint copy has its own icon", () => {
  const original = images.typeIcon(RIFTER_BLUEPRINT);
  const copy = images.typeIcon(RIFTER_BLUEPRINT, { copy: true });
  assert.equal(copy.length, 1);
  assert.notEqual(original[0].src, copy[0].src);
  assertExists(copy[0].src);

  assert.deepEqual(images.typeIcon(RIFTER, { copy: true }), images.typeIcon(RIFTER));
});

test("a module blueprint is framed, with an additive overlay", () => {
  const layers = images.typeIcon(DAMAGE_CONTROL_II_BLUEPRINT, { marker: false });
  assert.equal(layers.length, 3);
  assert.equal(layers[1].src, images.typeIcon(DAMAGE_CONTROL_II)[0].src);
  assert.equal(layers[2].additive, true);
  layers.forEach((layer) => assertExists(layer.src));

  const copy = images.typeIcon(DAMAGE_CONTROL_II_BLUEPRINT, { copy: true, marker: false });
  assert.notEqual(copy[0].src, layers[0].src);
  assert.equal(copy[1].src, layers[1].src);
});

test("a relic has no copy", () => {
  const layers = images.typeIcon(INTACT_HULL_SECTION);
  assert.equal(layers.length, 3);
  assert.deepEqual(images.typeIcon(INTACT_HULL_SECTION, { copy: true }), layers);
});

test("unknown things have no icon", () => {
  assert.equal(images.typeIcon(-1), undefined);
  assert.equal(images.typeIcon(999999999), undefined);
  assert.equal(images.marketGroupIcon(-1), undefined);
  assert.equal(images.metaGroupIcon(-1), undefined);
  assert.equal(images.uiTexture("does/not/exist"), undefined);
});

test("groups and UI textures", () => {
  assertExists(images.marketGroupIcon(SHIPS));
  assertExists(images.uiTexture("classes/fitting/statsicons/armorhp"));
});

test("a file that is not an images file is refused", async () => {
  await assert.rejects(loadImages({ bytes: new Uint8Array(16) }, { baseUrl: "/" }));
});
