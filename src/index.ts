import { ByteBuffer } from "flatbuffers";

import { Frame, type FrameImages, Images as RawImages, type KeyedImage, type TypeIcon, type UiTexture } from "./generated/images.js";

export type Source = { url: string | URL } | { bytes: Uint8Array };

/** Layers are drawn at their own size, from the top-left corner of a 64 by 64 pixel icon. */
export interface Layer {
  src: string;
  /** Draw additively; in CSS, `mix-blend-mode: plus-lighter`. */
  additive?: true;
}

export interface TypeIconOptions {
  /** The icon of a blueprint copy; ignored for other types. */
  copy?: boolean;
  /** Include the tech level or faction marker; defaults to true. */
  marker?: boolean;
}

export interface ImagesOptions {
  /** Where this package's dist/images/ folder is served from. */
  baseUrl: string | URL;
}

const NO_MARKER = 0xff;

export class Images {
  readonly #raw: RawImages;
  readonly #baseUrl: string;
  readonly #urls: (string | undefined)[] = [];

  constructor(bytes: Uint8Array, options: ImagesOptions) {
    const buffer = new ByteBuffer(bytes);
    if (!RawImages.bufferHasIdentifier(buffer)) {
      throw new Error("Not an images file: the identifier is missing");
    }
    this.#raw = RawImages.getRootAsImages(buffer);

    const baseUrl = options.baseUrl.toString();
    this.#baseUrl = baseUrl.endsWith("/") ? baseUrl : `${baseUrl}/`;
  }

  /** The layers, bottom first. */
  typeIcon(typeId: number, options: TypeIconOptions = {}): Layer[] | undefined {
    const icon = search(this.#raw.typesLength(), (index) => this.#raw.types(index)!, (entry: TypeIcon) => entry.typeId() - typeId);
    if (icon === undefined) return undefined;

    const layers: Layer[] = [];
    const frame = this.#frame(icon.frame(), options.copy === true);
    if (frame === undefined) {
      const copy = options.copy ? this.#keyed(this.#raw.blueprintCopiesLength(), (index) => this.#raw.blueprintCopies(index)!, typeId) : undefined;
      layers.push({ src: this.#url(copy ?? icon.icon()) });
    } else {
      layers.push({ src: this.#url(frame.background()) }, { src: this.#url(icon.icon()) }, { src: this.#url(frame.overlay()), additive: true });
    }

    const marker = icon.marker();
    if (options.marker !== false && marker !== NO_MARKER) layers.push({ src: this.#url(this.#raw.metaGroups(marker)!.image()) });
    return layers;
  }

  marketGroupIcon(marketGroupId: number): string | undefined {
    const image = this.#keyed(this.#raw.marketGroupsLength(), (index) => this.#raw.marketGroups(index)!, marketGroupId);
    return image === undefined ? undefined : this.#url(image);
  }

  /** Also the marker of the meta group's types. */
  metaGroupIcon(metaGroupId: number): string | undefined {
    const image = this.#keyed(this.#raw.metaGroupsLength(), (index) => this.#raw.metaGroups(index)!, metaGroupId);
    return image === undefined ? undefined : this.#url(image);
  }

  factionIcon(factionId: number): string | undefined {
    const image = this.#keyed(this.#raw.factionsLength(), (index) => this.#raw.factions(index)!, factionId);
    return image === undefined ? undefined : this.#url(image);
  }

  attributeIcon(attributeId: number): string | undefined {
    const image = this.#keyed(this.#raw.attributesLength(), (index) => this.#raw.attributes(index)!, attributeId);
    return image === undefined ? undefined : this.#url(image);
  }

  effectIcon(effectId: number): string | undefined {
    const image = this.#keyed(this.#raw.effectsLength(), (index) => this.#raw.effects(index)!, effectId);
    return image === undefined ? undefined : this.#url(image);
  }

  /** By its path below res:/ui/texture/, without extension. */
  uiTexture(name: string): string | undefined {
    const texture = search(this.#raw.uiTexturesLength(), (index) => this.#raw.uiTextures(index)!, (entry: UiTexture) => compare(entry.name()!, name));
    return texture === undefined ? undefined : this.#url(texture.image());
  }

  #frame(frame: Frame, copy: boolean): FrameImages | undefined {
    switch (frame) {
      case Frame.Blueprint:
        return copy ? this.#raw.blueprintCopy()! : this.#raw.blueprint()!;
      case Frame.Relic:
        return this.#raw.relic()!;
      case Frame.Reaction:
        return this.#raw.reaction()!;
      default:
        return undefined;
    }
  }

  #keyed(length: number, at: (index: number) => KeyedImage, id: number): number | undefined {
    return search(length, at, (entry) => entry.id() - id)?.image();
  }

  #url(index: number): string {
    let url = this.#urls[index];
    if (url === undefined) {
      const hash = this.#raw.images(index)!.hash();
      url = `${this.#baseUrl}${hash.toString(16).padStart(16, "0")}.webp`;
      this.#urls[index] = url;
    }
    return url;
  }
}

export async function loadImages(source: Source, options: ImagesOptions): Promise<Images> {
  return new Images(await readSource(source), options);
}

async function readSource(source: Source): Promise<Uint8Array> {
  if ("bytes" in source) return source.bytes;

  const response = await fetch(source.url);
  if (!response.ok) {
    throw new Error(`Fetching ${source.url.toString()} failed: ${response.status} ${response.statusText}`);
  }
  return new Uint8Array(await response.arrayBuffer());
}

function search<T>(length: number, at: (index: number) => T, compareTo: (entry: T) => number): T | undefined {
  let low = 0;
  let high = length - 1;
  while (low <= high) {
    const middle = (low + high) >>> 1;
    const entry = at(middle);
    const order = compareTo(entry);
    if (order === 0) return entry;
    if (order < 0) low = middle + 1;
    else high = middle - 1;
  }
  return undefined;
}

// Go sorts the names on bytes; for ASCII that is the same order as JavaScript's.
function compare(a: string, b: string): number {
  return a < b ? -1 : a > b ? 1 : 0;
}
