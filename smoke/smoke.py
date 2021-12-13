from dataclasses import dataclass

import object_graph_streamer as ogs

@dataclass
class Test:
  Yoo: int
  Bla: int


sample ={"Yoo":9, "Bla": 5}
out = []
jsonC = ogs.JsonCollector(lambda o: out.append(o))
ogs.objectGraphStreamer(sample, lambda prob: jsonC.append(prob))
if "".join(out) != "{\"Bla\":5,\"Yoo\":9}":
  raise Exception(f'out={out}')

print("Ready for production")
