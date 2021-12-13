import * as osg from "object-graph-streamer";


const sample = {Yoo: 9, Bla: 5}
let out = ""
const jsonC = new osg.JsonCollector((o: string) => { out += o })
osg.objectGraphStreamer(sample, (prob: osg.SVal) => jsonC.append(prob))
if (out != "{\"Bla\":5,\"Yoo\":9}") {
  throw Error(`out=${out}`);
}
console.log("Ready for production");

