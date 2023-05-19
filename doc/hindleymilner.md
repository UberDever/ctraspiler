Constraints

NodeIntLiteral {token}: [node] = int; push [node]
NodeIdentifier {token}: [node]; push [node]
NodeBinaryPlus {lhs rhs}: pop [lhs]; pop [rhs]; [node] = [lhs] = [rhs]