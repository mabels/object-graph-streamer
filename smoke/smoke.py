from dataclasses import dataclass

from object_graph_streamer as ogs

@dataclass
class Test:
  Yoo: int
  Bla: int

sample = Test(Yoo=9, Bla= 5)
out = ""
jsonC = ogs.NewJsonCollector(lambda o: out += o)
ogs.ObjectGraphStreamer(sample, lambda prob: jsonC.Append(prob))
if out != "{\"Bla\":5,\"Yoo\":9}":
  raise Exception(f'out={out}')

print("Ready for production")
