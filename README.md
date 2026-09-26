# Images

[![npm](https://img.shields.io/npm/v/%40eveshipfit%2Fimages.svg)](https://www.npmjs.com/package/@eveshipfit/images)
[![CI](https://github.com/EVEShipFit/images/actions/workflows/testing.yml/badge.svg)](https://github.com/EVEShipFit/images/actions/workflows/testing.yml)

[![Discord](https://img.shields.io/badge/Discord-Join-5865F2?logo=discord&logoColor=white)](https://discord.gg/S5V5BkvNf7)

EVE Online's icons, as WebP.

Every published type, market group and meta group has its icon here, together with the client's UI textures EVEShipFit uses.
Icons are not composited: a type has an icon, a frame and a marker, so the tech level or faction marker can be left off.

## Why?

Most images are also available on https://images.evetech.net. But these images aren't
always of the best quality: sometimes their marker is wrong, etc. See [here](https://github.com/esi/esi-issues/issues/1448) for an overview
of issues.

Additionally, in-game icons aren't available on that server. Meaning we have to fetch
them from the client resources.

Both combined created this repository: it fixes both issues by just fetching all the
information from the client data directly. Additionally, it outputs it as WebP with
a quality of 0.9, which makes the files about 5 times smaller.

## Prerequisite

- Go 1.27 or later
- flatc, to change the schema
- Node, for the npm package
- libwebp, optional; converting is several times faster with it

## Usage

```bash
go run ./cmd/images download        # fetch the latest SDE into sde/, and the client's resource index into cache/
go run ./cmd/images build           # convert the images, and write them with images.dat to dist/
go run ./cmd/images compare old/    # tell whether dist/ differs from old/
```

`build` downloads the SDE when there is none yet, and uses the one on disk otherwise.
Use `--build <number>` to pin a version; the client shares its build number with the SDE.

Converted images are kept in `cache/`, so a next build only fetches what changed in EVE.

## Output

`build` writes to `dist/`:
- `images/`, one WebP per image. A file is named after a hash of its source, so it keeps its name while it does not change.
- `images.dat`, a flatbuffer, described in [specs/](specs/images.fbs), which tells which images every type, market group and meta group uses.

Icons are scaled to 64 by 64 pixels, lossy at quality 90.
Icons smaller than that, markers, frames and UI textures keep their own size, and are lossless.

Which image a type gets follows what the client draws, as worked out by [EVE-TurtleTools](https://github.com/SentientTurtle/EVE-TurtleTools)' eveicongenerator.

## npm

```sh
npm install @eveshipfit/images
```

The package ships `dist/images.dat`, the `dist/images/` folder and a loader.
Serve the folder with your site, and load `images.dat` like the SDE.
With Vite and [vite-plugin-static-copy](https://github.com/sapphi-red/vite-plugin-static-copy):

```js
// vite.config.js
viteStaticCopy({
  targets: [{ src: "node_modules/@eveshipfit/images/dist/images/*", dest: "images", rename: { stripBase: true } }],
});
```

```ts
import imagesUrl from "@eveshipfit/images/dist/images.dat?url";
import { loadImages } from "@eveshipfit/images";

const images = await loadImages({ url: imagesUrl }, { baseUrl: `${import.meta.env.BASE_URL}images/` });

images.typeIcon(2048); // Damage Control II: [{ src }, { src }], the icon and its Tech II marker
images.typeIcon(2048, { marker: false }); // without the Tech II marker
images.typeIcon(691, { copy: true }); // Rifter Blueprint, as a copy
images.marketGroupIcon(4); // Ships
images.metaGroupIcon(2); // Tech II
images.uiTexture("classes/fitting/statsicons/armorhp");
```

A type's layers are drawn at their own size, from the top-left corner of an icon of 64 pixels; a marker is smaller than the icon below it.
Scale the whole icon rather than the layers.
An `additive` layer is added onto the layers below; draw the stack in its own stacking context, so it is not added onto the page:

```tsx
function TypeIcon({ typeId, size = 64 }: { typeId: number; size?: number }) {
  const layers = images.typeIcon(typeId);
  if (layers === undefined) return null;

  return (
    <span style={{ position: "relative", display: "inline-block", width: 64, height: 64, zoom: size / 64, isolation: "isolate" }}>
      {layers.map((layer) => (
        <img
          key={layer.src}
          src={layer.src}
          alt=""
          style={{ position: "absolute", top: 0, left: 0, mixBlendMode: layer.additive ? "plus-lighter" : undefined }}
        />
      ))}
    </span>
  );
}
```

## Releasing

Every day at 12:00 UTC, `main` is built against the latest SDE and client.
When `images.dat` differs from the latest release, a new release is made and published on npm.

## License

All EVE Online data belongs to CCP, and is subject to [their license agreement](https://developers.eveonline.com/license-agreement).
Everything EVEShip.fit-specific is MIT.
