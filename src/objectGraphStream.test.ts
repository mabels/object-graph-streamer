import {
  objectGraphStreamer,
  JsonCollector,
  HashCollector,
} from "./objectGraphStreamer";

beforeAll(() => {
  // Lock Time
  jest.useFakeTimers("modern");
  jest.setSystemTime(new Date(1624140000000));
  // jest.spyOn(Date, 'now').mockImplementation(() => 1487076708000);
});

it("test simple hash", () => {
  // expect.assertions(3);

  const hashC = new HashCollector();
  objectGraphStreamer(
    {
      kind: "test",
      data: {
        name: "object",
        date: "2021-05-20",
      },
    },
    (sval) => hashC.append(sval)
  );
  expect(hashC.digest()).toBe("5zWhdtvKuGob1FbW9vUGPQKobcLtYYr5wU8AxQRVraeB");
});

it("sort with out with string", () => {
  const fn = jest.fn();
  objectGraphStreamer("string", fn);
  expect(fn.mock.calls).toEqual([
    [{ val: { val: "string" }, path: [], outState: "Value" }],
  ]);
});

it("sort with out with date", () => {
  const fn = jest.fn();
  objectGraphStreamer(new Date(444), fn);
  expect(fn.mock.calls).toEqual([
    [{ val: { val: new Date(444) }, path: [], outState: "Value" }],
  ]);
});

it("sort with out with number", () => {
  const fn = jest.fn();
  objectGraphStreamer(4711, fn);
  expect(fn.mock.calls).toEqual([
    [{ val: { val: 4711 }, path: [], outState: "Value" }],
  ]);
});

it("sort with out with boolean", () => {
  const fn = jest.fn();
  objectGraphStreamer(false, fn);
  expect(fn.mock.calls).toEqual([
    [{ val: { val: false }, path: [], outState: "Value" }],
  ]);
});

it("sort with out with array of empty", () => {
  const fn = jest.fn();
  objectGraphStreamer([], fn);
  expect(fn.mock.calls).toEqual([
    [{ outState: "[", path: ["["] }],
    [{ outState: "]", path: ["]"] }],
  ]);
});

it("sort with out with array of [1,2]", () => {
  const fn = jest.fn();
  objectGraphStreamer([1, 2], fn);
  expect(fn.mock.calls).toEqual([
    [{ outState: "[", path: ["["] }],
    [{ val: { val: 1 }, outState: "Value", path: ["[", "0"] }],
    [{ val: { val: 2 }, outState: "Value", path: ["[", "1"] }],
    [{ outState: "]", path: ["]"] }],
  ]);
});

it("sort with out with array of [[1,2],[3,4]]", () => {
  const fn = jest.fn();
  objectGraphStreamer(
    [
      [1, 2],
      [3, 4],
    ],
    fn
  );
  expect(fn.mock.calls).toEqual([
    [{ outState: "[", path: ["["] }],
    [{ outState: "[", path: ["[", "0", "["] }],
    [{ val: { val: 1 }, outState: "Value", path: ["[", "0", "[", "0"] }],
    [{ val: { val: 2 }, outState: "Value", path: ["[", "0", "[", "1"] }],
    [{ outState: "]", path: ["[", "0", "]"] }],
    [{ outState: "[", path: ["[", "1", "["] }],
    [{ val: { val: 3 }, outState: "Value", path: ["[", "1", "[", "0"] }],
    [{ val: { val: 4 }, outState: "Value", path: ["[", "1", "[", "1"] }],
    [{ outState: "]", path: ["[", "1", "]"] }],
    [{ outState: "]", path: ["]"] }],
  ]);
});

it("sort with out with obj of {}", () => {
  const fn = jest.fn();
  objectGraphStreamer({}, fn);
  expect(fn.mock.calls).toEqual([
    [{ outState: "{", path: ["{"] }],
    [{ outState: "}", path: ["}"] }],
  ]);
});

it("sort with out with obj of { y: 1, x: 2 }", () => {
  const fn = jest.fn();
  objectGraphStreamer({ y: 1, x: 2 }, fn);
  expect(fn.mock.calls).toEqual([
    [{ outState: "{", path: ["{"] }],
    [{ attribute: "x", outState: "Attr", path: ["{", "x"] }],
    [{ val: { val: 2 }, outState: "Value", path: ["{", "x"] }],
    [{ attribute: "y", outState: "Attr", path: ["{", "y"] }],
    [{ val: { val: 1 }, outState: "Value", path: ["{", "y"] }],
    [{ outState: "}", path: ["}"] }],
  ]);
});

it("sort with out with obj of { y: { b: 1, a: 2 }  }", () => {
  const fn = jest.fn();
  objectGraphStreamer({ y: { b: 1, a: 2 } }, fn);
  expect(fn.mock.calls).toEqual([
    [{ outState: "{", path: ["{"] }],
    [{ attribute: "y", outState: "Attr", path: ["{", "y"] }],
    [{ outState: "{", path: ["{", "y", "{"] }],
    [{ attribute: "a", outState: "Attr", path: ["{", "y", "{", "a"] }],
    [{ val: { val: 2 }, outState: "Value", path: ["{", "y", "{", "a"] }],
    [{ attribute: "b", outState: "Attr", path: ["{", "y", "{", "b"] }],
    [{ val: { val: 1 }, outState: "Value", path: ["{", "y", "{", "b"] }],
    [{ outState: "}", path: ["{", "y", "}"] }],
    [{ outState: "}", path: ["}"] }],
  ]);
});

it("JSONCollector {}", () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o));
  objectGraphStreamer({}, (o) => json.append(o));
  expect(out).toBe("{}");
});
it("JSONCollector []", () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o));
  objectGraphStreamer([], (o) => json.append(o));
  expect(out).toBe("[]");
});

it('JSONCollector { x: { y: 1, z: "x" }, y: {}, z: []}', () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o));
  objectGraphStreamer({ x: { y: 1, z: "x" }, y: {}, z: [] }, (o) =>
    json.append(o)
  );
  expect(out).toBe('{"x":{"y":1,"z":"x"},"y":{},"z":[]}');
});

it('JSONCollector ["xx"]', () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o));
  objectGraphStreamer(["xx"], (o) => json.append(o));
  expect(out).toBe('["xx"]');
});

it('JSONCollector [1, "2"]', () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o));
  objectGraphStreamer([1, "2"], (o) => json.append(o));
  expect(out).toBe('[1,"2"]');
});

it('JSONCollector [1, ["2", "A"]]', () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o));
  objectGraphStreamer([1, ["2", "A"], "E"], (o) => json.append(o));
  expect(out).toBe('[1,["2","A"],"E"]');
});

it("JSONCollector indent=2 {} ", () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o), { indent: 2 });
  objectGraphStreamer({}, (o) => json.append(o));
  expect(out).toBe("{}");
});

it("JSONCollector indent=2 [] ", () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o), { indent: 2 });
  objectGraphStreamer([], (o) => json.append(o));
  expect(out).toBe("[]");
});

it('JSONCollector indent=2 { x: { y: 1, z: "x" }}', () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o), { indent: 2 });
  objectGraphStreamer({ x: { y: 1, z: "x" }, y: {}, z: [] }, (o) =>
    json.append(o)
  );
  expect(out).toBe(
    '{\n  "x": {\n    "y": 1,\n    "z": "x"\n  },\n  "y": {},\n  "z": []\n}'
  );
});

it('JSONCollector indent=2 ["xx"]', () => {
  let out = "";
  const json = new JsonCollector((o) => (out += o), { indent: 2 });
  objectGraphStreamer(["xx"], (o) => json.append(o));
  expect(out).toBe('[\n  "xx"\n]');
});

it('JSONCollector indent=2 [1, "2"]', () => {
  let out = "";
  const json = new JsonCollector(
    (o) => {
      // console.log("OUT=>", o)
      out += o;
    },
    { indent: 2 }
  );
  objectGraphStreamer([1, "2"], (o) => json.append(o));
  expect(out).toBe('[\n  1,\n  "2"\n]');
});

it("JSONCollector [1, new Date(444)]", () => {
  let out = "";
  const json = new JsonCollector((o) => {
    // console.log("OUT=>", o)
    out += o;
  });
  const obj = [1, new Date(444)];
  objectGraphStreamer(obj, (o) => json.append(o));
  expect('[1,"1970-01-01T00:00:00.444Z"]').toBe(out);
});

it("HashCollector 1", () => {
  const hash = new HashCollector();
  objectGraphStreamer(
    { x: { y: 1, z: "x" }, y: {}, z: [], d: new Date(444) },
    (o) => hash.append(o)
  );
  expect(hash.digest()).toBe("5PvJAWGkaKAHax6tsaKGfPYm6JfXxZs15wRTDpSKaZ2G");
});

it("HashCollector 2", () => {
  const hash = new HashCollector();
  objectGraphStreamer(
    { x: { y: 2, z: "x" }, y: {}, z: [], date: new Date(444) },
    (o) => hash.append(o)
  );
  expect(hash.digest()).toBe("ECVWfmcNaUGkgvPZe7CojrnRNULxNczKXU8PGns6UDvr");
});

it("HashCollector 3", () => {
  const hash = new HashCollector();
  objectGraphStreamer(
    { x: { x: 1, z: "x" }, y: {}, z: [], date: new Date(444) },
    (o) => hash.append(o)
  );
  expect(hash.digest()).toBe("EoYNGMtap1k9iEAGeVtHmJwpMjQLKWJmR27SG6aC9fSg");
});

it("HashCollector 4", () => {
  const hash1 = new HashCollector();
  objectGraphStreamer(
    { x: { x: 1, z: "x" }, y: {}, z: [], date: new Date(444) },
    (o) => hash1.append(o)
  );
  const hash2 = new HashCollector();
  objectGraphStreamer(
    { date: new Date(444), x: { x: 1, z: "x" }, y: {}, z: [] },
    (o) => hash2.append(o)
  );
  expect(hash1.digest()).toBe(hash2.digest());
});

it("HashCollector 3 internal update", () => {
  const hashCollector = new HashCollector();
  hashCollector.hash.update = jest.fn();
  objectGraphStreamer(
    { x: { r: 1, z: "u" }, y: {}, z: [], date: new Date(444) },
    (o) => hashCollector.append(o)
  );
  const result = (hashCollector.hash.update as jest.Mock).mock.calls.map(
    ([elem]: Buffer[]) => elem.toString()
  );
  expect(result).toEqual([
    "date",
    "1970-01-01T00:00:00.444Z",
    "x",
    "r",
    "1",
    "z",
    "u",
    "y",
    "z",
  ]);
  expect(hashCollector.digest()).toBe(
    "GKot5hBsd81kMupNCXHaqbhv3huEbxAFMLnpcX2hniwn"
  );
});

afterAll(() => {
  // Unlock Time
  jest.useRealTimers();
});
