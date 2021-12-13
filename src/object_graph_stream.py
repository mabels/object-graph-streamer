import copy
from dataclasses import dataclass

import typing

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


idGeneratorFN = typing.Callable[[any], None]


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
    NONE = "NE",
    ARRAY_START = "[",
    ARRAY_END = "]",
    OBJECT_START = "{",
    OBJECT_END = "}",


class SVal:
    attribute: str
    val: any
    outState: OutState

    def __init__(self, attribute=None, val=None, outState=None) -> None:
        self.attribute = attribute
        self.val = val
        self.outState = outState

    def to_dict(self):
        ret = {}
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
        if sval.outState is not None:
            return
        if sval.attribute is not None:
            tmp = sval.attribute.encode('utf-8')
            # print("attribute=", tmp)
            self.hash.update(tmp)
        if sval.val is not None:
            out = sval.val.asValue()
            if isinstance(out, datetime):
                out = jsIsoFormat(out)
            else:
                out = str(out)
            # print("val=", out)
            self.hash.update(out.encode("utf-8"))


def objectGraphStreamer(e: any, out: typing.Callable[[SVal], None]):
    if isinstance(e, list):
        out(SVal(**{'outState': OutState.ARRAY_START}))
        for i in e:
            objectGraphStreamer(i, out)
        out(SVal(**{'outState': OutState.ARRAY_END}))
        return
    elif isinstance(e, dict):
        out(SVal(**{'outState': OutState.OBJECT_START}))
        keys = list(e.keys())
        keys.sort()
        # print("keys=", keys)
        for i in keys:
            out(SVal(**{'attribute': i}))
            objectGraphStreamer(e[i], out)

        out(SVal(**{'outState': OutState.OBJECT_END}))
        return
    else:
        # print(f"VAL[{e}]")
        out(SVal(**{'val': JsonValType(e)}))
        return


@dataclass
class JsonHash:
    jsonStr: str
    hash: str

