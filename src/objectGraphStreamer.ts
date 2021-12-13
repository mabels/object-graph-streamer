import * as crypto from "crypto";
import baseX from "base-x";

const bs58 = baseX(
  "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
);

export interface JsonProps {
  readonly indent: number;
  readonly newLine: string;
}

type OutputFn = (str: string) => void;

export class JsonCollector {
  readonly output: OutputFn;
  readonly indent: string;
  readonly commas: string[] = [""];
  readonly elements: number[] = [0];
  readonly props: JsonProps;
  readonly nextLine: string;
  attribute?: string;

  constructor(output: OutputFn, props: Partial<JsonProps> = {}) {
    this.output = output;
    this.props = {
      indent: props.indent || 0,
      newLine: props.newLine || "\n",
      ...props,
    };
    this.indent = Array(this.props.indent).fill(" ").join("");
    this.nextLine = this.props.indent ? this.props.newLine : "";
  }

  public get suffix(): string {
    if (this.elements[this.elements.length - 1]) {
      return (
        this.nextLine +
        Array(this.commas.length - 1)
          .fill(this.indent)
          .join("")
      );
    } else {
      return "";
    }
  }

  public append(sval: SVal) {
    if (sval.outState) {
      switch (sval.outState) {
        case OutState.ARRAY_START:
          this.output(
            this.commas[this.commas.length - 1] +
              this.suffix +
              (this.attribute || "") +
              "["
          );
          this.attribute = undefined;
          this.commas[this.commas.length - 1] = ",";
          this.commas.push("");
          this.elements.push(0);
          break;

        case OutState.ARRAY_END:
          this.commas.pop();
          this.output(this.suffix + "]");
          this.elements.pop();
          break;

        case OutState.OBJECT_START:
          this.output(
            this.commas[this.commas.length - 1] +
              this.suffix +
              (this.attribute || "") +
              "{"
          );
          this.attribute = undefined;
          this.commas[this.commas.length - 1] = ",";
          this.commas.push("");
          this.elements.push(0);
          break;

        case OutState.OBJECT_END:
          this.commas.pop();
          this.output(this.suffix + "}");
          this.elements.pop();
          break;
      }
    }
    if (sval.val) {
      ++this.elements[this.elements.length - 1];
      const out =
        this.commas[this.commas.length - 1] +
        this.suffix +
        (this.attribute || "") +
        sval.val.toString();
      // console.log(this.commas, this.attribute, sval.val, out);
      this.output(out);
      this.attribute = undefined;
      this.commas[this.commas.length - 1] = ",";
      // }
    }
    if (sval.attribute) {
      ++this.elements[this.elements.length - 1];
      this.attribute =
        JSON.stringify(sval.attribute) + ":" + (this.indent.length ? " " : "");
    }
  }
}

export class HashCollector {
  readonly hash: crypto.Hash = crypto.createHash("sha256");

  constructor() {}

  public digest() {
    return bs58.encode(this.hash.digest());
  }

  public append(sval: SVal) {
    if (sval.outState === OutState.ATTRIBUTE && sval.attribute) {
      const tmp = Buffer.from(sval.attribute).toString('utf-8')
      // console.log('attribute=', tmp)
      this.hash.update(tmp);
    }
    if (sval.outState === OutState.VALUE && sval.val) {
      let out: any = sval.val.asValue();
      if (out instanceof Date) {
        out = out.toISOString()
      } else {
        out = "" + out;
      }
      // console.log('val=', out)
      // We need some room for the types
      this.hash.update(Buffer.from(out).toString("utf-8"));
    }
  }
}

export interface JsonHash {
  readonly jsonStr: string;
  readonly hash?: string;
}

export function lexicalSort(a: number | string, b: number | string): number {
  if (a < b) {
    return -1;
  }
  if (a > b) {
    return 1;
  }
  return 0;
}

type ValueType = string | number | boolean | Date | undefined;

export interface ValType {
  toString(): string;
  asValue(): any;
}

export class JsonValType implements ValType {
  readonly val: ValueType;

  constructor(val: ValueType) {
    this.val = val;
  }

  public asValue() {
    return this.val;
  }

  public toString() {
    return JSON.stringify(this.val);
  }
}

export class PlainValType implements ValType {
  readonly val: string;

  constructor(val: string) {
    this.val = val;
  }

  public asValue() {
    return this.val;
  }

  public toString() {
    return this.val;
  }
}

export enum OutState {
  VALUE = "Value",
  ATTRIBUTE = "Attr",
  ARRAY_START = "[",
  ARRAY_END = "]",
  OBJECT_START = "{",
  OBJECT_END = "}",
}

export interface SVal {
  readonly attribute?: string;
  readonly val?: ValType;
  readonly outState: OutState;
  readonly path: string[]
}

export type SValFn = (prob: SVal) => void;

export interface ObjectGraphStreamerProps {
    readonly paths: string[];
    readonly objectProcessor: (a: Array<string>) => Array<string>
    readonly arrayProcessor: (a: Array<unknown>) => Array<unknown>
    readonly valFactory: (e: any) => ValType
}

function defaultObjectGraphStreamerProps(ogsp?: Partial<ObjectGraphStreamerProps>): ObjectGraphStreamerProps {
  ogsp = ogsp || {}
  return {
    paths: ogsp.paths || [],
    objectProcessor: ogsp.objectProcessor || ((a: string[]) => { a.sort(lexicalSort); return a}),
    arrayProcessor: ogsp.arrayProcessor || ((a: Array<unknown>) => { return a; }),
    valFactory: ogsp.valFactory || ((e: any) => (new JsonValType(e))),
  }
}

export function objectGraphStreamer<T>(e: T, 
    streamFn: SValFn, 
    pogsp?: Partial<ObjectGraphStreamerProps>): void {
  const ogsp = defaultObjectGraphStreamerProps(pogsp) 
  if (Array.isArray(e)) {
    const arrayPaths = [...ogsp.paths, "["]
    streamFn({ outState: OutState.ARRAY_START, path:arrayPaths });
    ogsp.arrayProcessor(e).forEach((item, idx) => {
      objectGraphStreamer(item, streamFn, {
        ...ogsp,
        paths: [...arrayPaths, `${idx}`]
      })
    });
    streamFn({ outState: OutState.ARRAY_END, path: [...ogsp.paths, "]"] });
    return;
  } else if (typeof e === "object" && !(e instanceof Date)) {
    const objectPath = [...ogsp.paths, "{"]
    streamFn({ outState: OutState.OBJECT_START, path: objectPath  });
    ogsp.objectProcessor(Object.keys(e))
      .forEach((attribute) => {
        const attrPath = [...objectPath, attribute]
        streamFn({ attribute: attribute, path: attrPath, outState: OutState.ATTRIBUTE });
        objectGraphStreamer((e as any)[attribute], streamFn, {
          ...ogsp,
          paths: attrPath
        })
      });
    streamFn({ outState: OutState.OBJECT_END, path: [...ogsp.paths, "}"] });
    return;
  } else {
    streamFn({ val: ogsp.valFactory(e), path: ogsp.paths, outState: OutState.VALUE});
    return;
  }
}
