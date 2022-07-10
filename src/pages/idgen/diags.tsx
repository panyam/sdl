import * as React from "react"
import * as sdsim from "sdsim"

export const Diagram = (props: any) => {
return <div></div>
}

export const SimpleIDG = () => {
  const catalog = IDGenCatalog();
  const id1 = catalog.newInstance("IDGen", "idgen0");

  // how do we specify x/y and orderings?
  // Do we just keep this declarative and be auto-determined during
  // layouts - a simple BFS based layout may be good enough for now
  const diag = new NodeView("Simple MVP");
  const a1 = diag.nodeView("User", "actor.png")
  const c1 = diag.nodeView("Client", "host.png")
  diag.connect(a1, c1);
  const idv1 = diag.nodeView(id1);
  diag.connect(c1, idv1);
  return diag;
}
