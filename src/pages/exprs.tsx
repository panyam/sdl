import * as React from "react"
import * as dagmath from "dagmath"

function getDAG(): dagmath.DAG {
  const w = window as any
  if (!w.pageSystemDAG) {
    const dag = dagmath.StdLib.PopulateDAG(new dagmath.DAG());
    w.pageSystemDAG = dag;

    // set a couple of variables typically used
    // targetVar = dag.setValue(targetVar, expr);
    // targetVar.desc = props.desc || "";
  }
  return w.pageSystemDAG;
}

function getParser(): dagmath.Parser.Parser {
  const w = window as any
  if (!w.parser) {
    // Use JS operator precedence
    w.parser = new dagmath.Parser.Parser(getDAG()).setOP(...dagmath.Parser.JSOperators);
  }
  return w.parser;
}

export const VarNameDescNode = (props: any) => {
  const vname = props.var;
  const dag = getDAG();
  const targetVar = dag.getVar(vname);
  if (targetVar == null) {
    return (<span className="invalid_var_name">Invalid variable name: {vname}</span>)
  } else {
    return (<span className="var_desc_span">{targetVar.desc}, <VarNameNode>{targetVar.name}</VarNameNode> = <ExprNode>{targetVar.name}</ExprNode></span>)
  }
}

export const VarNameNode = (props: any) => {
    return (<span className="var_name_span"> {props.children} </span>)
}

export const CodeNode = (props: any) => {
  return (<pre><code>{props.children}</code></pre>)
}

export const ExprNode = (props: any) => {
  const dag = getDAG();
  const parser = getParser();
  const expr = parser.parse(props.children);
  if (props.setto) {
    let targetVar = props.setto;
    if (dag.getVar(targetVar) != null) {
      console.warn("Variable already assigned: ", targetVar);
    } else {
      targetVar = dag.setValue(targetVar, expr);
      targetVar.desc = props.desc || "";
    }
  }
  const hidden = (props.hidden || "").toLowerCase()
  if (hidden == "true") {
    return null;
  }
  const val = expr.latestValue;
  // Also render the value
  if (val == dag.NULL || !val) {
    return (<span className="param_desc 3"> InvalidExp: {props.children} </span>)
  } else {
    return (<span className="param_desc 3"> {val.value.toLocaleString('en-US')} </span>)
  }
}

