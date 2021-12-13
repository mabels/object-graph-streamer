from dataclasses import dataclass

import typing
from enum import Enum

import json
from datetime import datetime
import hashlib
from base58 import b58encode


class JsonProps:
    indent: int
    newLine: str

    def __init__(self, indent=None, newLine=None) -> None:
        self.indent = 0 if indent is None else indent
        self.newLine = "\n" if newLine is None else newLine


class ValType:
    def toString(self): str

    def asValue(self): any

    def to_dict(self): dict


def jsIsoFormat(val: datetime):
    isoStr = val.isoformat().split(".")
    return f'{isoStr[0]}.{isoStr[1][0:3]}Z'


class JsonValType(ValType):
    val: any

    def __init__(self, val: any):
        self.val = val

    def asValue(self):
        return self.val

    def toString(self):
        val = self.val
        if isinstance(self.val, float):
            if float(self.val) == int(self.val):
                val = int(self.val)
        elif isinstance(self.val, datetime):
            val = jsIsoFormat(self.val)
        return json.dumps(val)
        # except Exception as e:
        # print("XXXXXXXX[", self.val, e)

    def to_dict(self):
        return {
            'val': self.val
        }


class PlainValType(ValType):
    val: str

    def __init__(self, val: str):
        self.val = val

    def asValue(self):
        return self.val

    def toString(self):
        return self.val

    def to_dict(self):
        return {
            'val': self.val
        }


class OutState(Enum):
    ATTRIBUTE = "Attr"
    VALUE = "Value"
    ARRAY_START = "[",
    ARRAY_END = "]",
    OBJECT_START = "{",
    OBJECT_END = "}",


class SVal:
    attribute: str
    val: any
    outState: OutState
    paths: typing.List[str]

    def __init__(self, outState, paths, attribute=None, val=None) -> None:
        self.attribute = attribute
        self.val = val
        self.outState = outState
        self.paths = paths

    def to_dict(self):
        ret = {}
        if self.paths is not None:
            ret['paths'] = self.paths
        if self.attribute is not None:
            ret['attribute'] = self.attribute
        if self.val is not None:
            ret['val'] = self.val.to_dict()
        if self.outState is not None:
            ret['outState'] = self.outState.value[0]
        return ret


OutputFN = typing.Callable[[str], None]


class JsonCollector:
    output: OutputFN
    indent: str
    commas: list[str]
    elements: list[int]
    props: JsonProps
    nextLine: str
    attribute: str

    def __init__(self, output: OutputFN, props: JsonProps = JsonProps()):
        self.output = output
        self.props = props
        self.indent = (" " * self.props.indent)
        self.nextLine = self.props.newLine if self.props.indent > 0 else ""
        self.commas = [""]
        self.elements = [0]
        self.attribute = ""
        # print("JsonCollector::__init__")

    def suffix(self) -> str:
        if self.elements[-1] > 0:
            return self.nextLine + (self.indent * (len(self.commas) - 1))
        else:
            return ""

    def append(self, sval: SVal):
        # print(f"append:{sval.to_dict()}-{this.commas}")
        if sval.outState is not None:
            if sval.outState == OutState.ARRAY_START:
                # print(f"Array-Start:{this.commas}-{this.suffix()}-{this.attribute}")
                self.output(
                    self.commas[-1] +
                    self.suffix() +
                    (self.attribute if self.attribute else "") +
                    "["
                )
                self.attribute = None
                self.commas[-1] = ","
                self.commas.append("")
                self.elements.append(0)
                return
            if sval.outState == OutState.ARRAY_END:
                self.commas.pop()
                self.output(self.suffix() + "]")
                self.elements.pop()
                return
            if sval.outState == OutState.OBJECT_START:
                self.output(
                    self.commas[-1] +
                    self.suffix() +
                    (self.attribute if self.attribute is not None else "") +
                    "{"
                )
                self.attribute = None
                self.commas[-1] = ","
                self.commas.append("")
                self.elements.append(0)
                return
            if sval.outState == OutState.OBJECT_END:
                self.commas.pop()
                self.output(self.suffix() + "}")
                self.elements.pop()
                return

        if sval.val is not None:
            self.elements[-1] = self.elements[-1] + 1
            # print(f"---[{sval.val}]-[{this.commas[-1]}]suffix[{this.suffix()}]attribute[{this.attribute}]val[{sval.val.toString()}]")
            out = self.commas[-1] + self.suffix() + (
                self.attribute if self.attribute is not None else "") + sval.val.toString()
            self.output(out)
            self.attribute = None
            self.commas[-1] = ","
        if sval.attribute:
            self.elements[-1] = self.elements[-1] + 1
            self.attribute = json.dumps(
                sval.attribute) + ":" + (" " if len(self.indent) > 0 else "")


class HashCollector:
    # readonly hash: crypto.Hash = crypto.createHash("sha256");
    hash: any  # hashlib._Hash

    def __init__(self, hash=None) -> None:
        self.hash = hashlib.new('sha256') if hash is None else hash

    def digest(self):
        return b58encode(self.hash.digest()).decode()

    def append(self, sval: SVal):
        if sval.outState == OutState.ATTRIBUTE:
            tmp = sval.attribute.encode('utf-8')
            # print("attribute=", tmp)
            self.hash.update(tmp)
        elif sval.outState == OutState.VALUE:
            out = sval.val.asValue()
            if isinstance(out, datetime):
                out = jsIsoFormat(out)
            else:
                out = str(out)
            # print("val=", out)
            self.hash.update(out.encode("utf-8"))


@dataclass
class ObjectGraphStreamerProps:
    paths: typing.Optional[typing.List[str]] = None
    objectProcessor: typing.Optional[typing.Callable[[
        typing.List[str]], typing.List[str]]] = None
    arrayProcessor:  typing.Optional[typing.Callable[[
        typing.List[any]], typing.List[any]]] = None
    valFactory: typing.Optional[typing.Callable[[any], ValType]] = None

    def assignPath(self, paths: typing.List[str]):
        return ObjectGraphStreamerProps(paths=paths,
                                        objectProcessor=self.objectProcessor,
                                        arrayProcessor=self.arrayProcessor,
                                        valFactory=self.valFactory)


def defaultObjectGraphStreamerProps(ogsp: typing.Optional[ObjectGraphStreamerProps]) -> ObjectGraphStreamerProps:
    if ogsp is None:
        ogsp = ObjectGraphStreamerProps(**{})
    else:
        ogsp = ObjectGraphStreamerProps(
            paths=ogsp.paths,
            objectProcessor=ogsp.objectProcessor,
            arrayProcessor=ogsp.arrayProcessor,
            valFactory=ogsp.valFactory)
    if not isinstance(ogsp.paths, list):
        ogsp.paths = []
    if not callable(ogsp.objectProcessor):
        def sorter(a):
            a.sort()
            return a
        ogsp.objectProcessor = sorter
    if not callable(ogsp.arrayProcessor):
        ogsp.arrayProcessor = lambda a: a
    if not callable(ogsp.valFactory):
        ogsp.valFactory = lambda a: JsonValType(a)
    return ogsp


def objectGraphStreamer(e: any, out: typing.Callable[[SVal], None], pogsp: typing.Optional[ObjectGraphStreamerProps] = None):
    ogsp = defaultObjectGraphStreamerProps(pogsp)
    if isinstance(e, list):
        arrayPaths = ogsp.paths + ["["]
        out(SVal(**{'outState': OutState.ARRAY_START, 'paths': arrayPaths}))
        for idx, i in enumerate(ogsp.arrayProcessor(e)):
            objectGraphStreamer(
                i, out, ogsp.assignPath(arrayPaths + [f"{idx}"]))
        out(SVal(**{'outState': OutState.ARRAY_END,
            'paths': ogsp.paths + [']']}))
        return
    elif isinstance(e, dict):
        attrPath = ogsp.paths + ['{']
        out(SVal(**{'outState': OutState.OBJECT_START, 'paths': attrPath}))
        for i in ogsp.objectProcessor(list(e.keys())):
            myPath = attrPath + [i]
            out(SVal(**{'attribute': i, 'paths': myPath,
                'outState': OutState.ATTRIBUTE}))
            objectGraphStreamer(e[i], out, ogsp.assignPath(myPath))

        out(SVal(**{'outState': OutState.OBJECT_END,
            'paths': ogsp.paths + ['}']}))
        return
    else:
        # print(f"VAL[{e}]")
        out(SVal(**{'val': JsonValType(e),
            'outState': OutState.VALUE, 'paths': ogsp.paths}))
        return
