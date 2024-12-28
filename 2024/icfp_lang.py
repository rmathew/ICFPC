#!/usr/bin/env python3
import argparse
import code
import logging
import math
import readline

DEC_KEY = list("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!\"#$%&'()*+,-./:;<=>?@[\]^_`|~ \n")
CHARSET_DELTA = 33
INT_BASE = 94
MAX_BETA_REDUCTIONS = 10000000
# MAX_BETA_REDUCTIONS = 5  # DEBUG
MIN_FAKE_VAR_NUM = 100


def _str_to_num(the_str):
    num = 0
    for c in the_str:
        num = num * INT_BASE + ord(c) - CHARSET_DELTA
    return num


def _cotbv_to_ascii(the_str):
    val = []
    for c in the_str:
        val += DEC_KEY[ord(c) - CHARSET_DELTA]
    return val


class Icfp(object):
    
    def __init__(self, tok):
        self.tok = tok

    def __str__(self):
        return self.tok

    def _dbg_out(self, depth, level_tracker, seq_out, tree_out):
        seq_out.append(f" {self.tok}")
        for i in range(depth):
            if i == depth - 1:
                tree_out.append("+- ")
            else:
                tree_out.append("|  " if level_tracker.get(i) is not None
                                else "   ")
        tree_out.append(str(self) + "\n")

    def token(self):
        return self.tok

    def substitute(self, var_id, replacement):
        return self

    def eval(self):
        raise NotImplementedError("Not implemented")

    @staticmethod
    def get_dbg_out(expr):
        level_tracker = {}
        seq_out = []
        tree_out = []
        expr._dbg_out(0, level_tracker, seq_out, tree_out)
        seq = "".join(seq_out)
        tree = "".join(tree_out)
        return f"*** STR ***\n {seq}\n*** AST:*** \n{tree}"

    @staticmethod
    def parse_token(tok, rest):
        if len(tok) < 1:
            raise SyntaxError("Invalid token for parsing")
        ind = tok[0]
        if ind == 'T' or ind == 'F':
            return IcfpBoolean._parse_token(tok)
        elif ind == 'I':
            return IcfpInteger._parse_token(tok)
        elif ind == 'S':
            return IcfpString._parse_token(tok)
        elif ind == 'U':
            return IcfpUnaryOp._parse_token(tok, rest)
        elif ind == 'B':
            return IcfpBinaryOp._parse_token(tok, rest)
        elif ind == '?':
            return IcfpConditional._parse_token(tok, rest)
        elif ind == 'L':
            return IcfpLambda._parse_token(tok, rest)
        elif ind == 'v':
            return IcfpVariable._parse_token(tok)
        else:
            raise SyntaxError(f"Unhandled indicator '{ind}'")


class IcfpBoolean(Icfp):

    def __init__(self, tok, val):
        Icfp.__init__(self, tok)
        self.val = val

    def eval(self):
        return self.val

    @staticmethod
    def _parse_token(tok):
        if tok[0] == 'T':
            val = True
        elif tok[0] == 'F':
            val = False
        else:
            raise SyntaxError("Invalid boolean literal")
        return IcfpBoolean(tok, val)


class IcfpInteger(Icfp):

    def __init__(self, tok, val):
        Icfp.__init__(self, tok)
        self.val = val

    def __str__(self):
        return f"{self.tok}  (#{self.val})"

    def eval(self):
        return self.val

    @staticmethod
    def _parse_token(tok):
        if tok[0] != 'I':
            raise SyntaxError("Invalid integer literal")
        body = tok[1:]
        if len(body) == 0:
            raise SyntaxError("Missing base-94 number in integer")
        val = _str_to_num(body)
        return IcfpInteger(tok, val)


class IcfpString(Icfp):

    def __init__(self, tok, val):
        Icfp.__init__(self, tok)
        self.val = val

    def __str__(self):
        val_str = "".join(self.val)
        return f"{self.tok}  (\"{val_str:.16}\")"

    def eval(self):
        return self.val

    @staticmethod
    def _parse_token(tok):
        if tok[0] != 'S':
            raise SyntaxError("Invalid string literal")
        val = _cotbv_to_ascii(tok[1:])
        return IcfpString(tok, val)


class IcfpUnaryOp(Icfp):

    def __init__(self, tok, operator, operand):
        Icfp.__init__(self, tok)

        if len(operator) != 1:
            raise SyntaxError("Malformed unary operator")
        if operator not in ['-', '!', '#', '$']:
            raise SyntaxError(f"Unknown unary operator '{operator}'")
        if operator == '-' and not isinstance(operand, IcfpInteger):
            raise SyntaxError("UnaryOp U- needs an integer operand")
        if operator == '!' and not isinstance(operand, IcfpBoolean):
            raise SyntaxError("UnaryOp U! needs a boolean operand")
        if operator == '#' and not isinstance(operand, IcfpString):
            raise SyntaxError("UnaryOp U# needs a string operand")
        if operator == '$' and not isinstance(operand, IcfpInteger):
            raise SyntaxError("UnaryOp U$ needs an integer operand")

        self.operator = operator
        self.operand = operand

    def _dbg_out(self, depth, level_tracker, seq_out, tree_out):
        super()._dbg_out(depth, level_tracker, seq_out, tree_out)
        self.operand._dbg_out(depth + 1, level_tracker, seq_out, tree_out)

    def substitute(self, var_id, replacement):
        new_operand = self.operand.substitute(var_id, replacement)
        if new_operand is self.operand:
            return self
        return IcfpUnaryOp(self.token(), self.operator, new_operand)

    def eval(self):
        if self.operator == '-':
            return -1 * self.operand.eval()
        elif self.operator == '!':
            return not self.operand.eval()
        elif self.operator == '#':
            return _str_to_num(self.operand.token()[1:])
        elif self.operator == '$':
            return _cotbv_to_ascii(self.operand.token()[1:])

    @staticmethod
    def _parse_token(tok, rest):
        if tok[0] != 'U':
            raise SyntaxError("Invalid unary operator")
        operator = tok[1:]
        if len(rest) < 1:
            raise SyntaxError("Missing operand for unary operator")
        operand = Icfp.parse_token(rest.pop(0), rest)
        return IcfpUnaryOp(tok, operator, operand)


class IcfpBinaryOp(Icfp):

    def __init__(self, tok, operator, operand1, operand2):
        Icfp.__init__(self, tok)

        if operator not in ['+', '-', '*', '/', '%', '<', '>', '=', '|',
                            '&', '.', 'T', 'D', '$']:
            raise SyntaxError(f"Unknown binary operator '{operator}'")
        self.operator = operator
        self.operand1 = operand1
        self.operand2 = operand2

    def _dbg_out(self, depth, level_tracker, seq_out, tree_out):
        super()._dbg_out(depth, level_tracker, seq_out, tree_out)
        level_tracker[depth] = True
        self.operand1._dbg_out(depth + 1, level_tracker, seq_out, tree_out)
        level_tracker.pop(depth, None)
        self.operand2._dbg_out(depth + 1, level_tracker, seq_out, tree_out)

    def substitute(self, var_id, replacement):
        new_operand1 = self.operand1.substitute(var_id, replacement)
        new_operand2 = self.operand2.substitute(var_id, replacement)
        if new_operand1 is self.operand1 and new_operand2 is self.operand2:
            return self
        return IcfpBinaryOp(self.token(), self.operator, new_operand1,
                            new_operand2)

    def eval(self):
        if self.operator == '+':
            return self.operand1.eval() + self.operand2.eval()
        if self.operator == '-':
            return self.operand1.eval() - self.operand2.eval()
        if self.operator == '*':
            return self.operand1.eval() * self.operand2.eval()
        if self.operator == '/':
            return int(float(self.operand1.eval()) /
                       float(self.operand2.eval()))
        if self.operator == '%':
            return int(math.fmod(float(self.operand1.eval()),
                                 float(self.operand2.eval())))
        if self.operator == '<':
            return self.operand1.eval() < self.operand2.eval()
        if self.operator == '>':
            return self.operand1.eval() > self.operand2.eval()
        if self.operator == '=':
            return self.operand1.eval() == self.operand2.eval()
        if self.operator == '|':
            return self.operand1.eval() or self.operand2.eval()
        if self.operator == '&':
            return self.operand1.eval() and self.operand2.eval()
        if self.operator == '.':
            return self.operand1.eval() + self.operand2.eval()
        if self.operator == 'T':
            return self.operand2.eval()[:self.operand1.eval()]
        if self.operator == 'D':
            return self.operand2.eval()[self.operand1.eval():]
        if self.operator == '$':
            if isinstance(self.operand1, IcfpLambda):
                if logging.getLogger().isEnabledFor(logging.DEBUG):
                    logging.debug(
                        f"Before beta-reduction:\n{Icfp.get_dbg_out(self)}")
                # A beta-reduction.
                beta_reduced_expr = self.operand1.ref_expr().substitute(
                    self.operand1.ref_var_id(), self.operand2)
                if logging.getLogger().isEnabledFor(logging.DEBUG):
                    logging.debug(f"After beta-reduction:\n"
                                  f"{Icfp.get_dbg_out(beta_reduced_expr)}")
                return beta_reduced_expr.eval()
            else:
                new_operand1 = self.operand1.eval()
                new_self = IcfpBinaryOp(self.token(), self.operator,
                                        new_operand1, self.operand2)
                return new_self.eval()
        raise ValueError(f"Unexpected binary operator '{self.operator}'")

    @staticmethod
    def _parse_token(tok, rest):
        if tok[0] != 'B':
            raise SyntaxError("Invalid binary operator")
        operator = tok[1:]
        if len(operator) != 1:
            raise SyntaxError("Malformed binary operator")
        if len(rest) < 2:
            raise SyntaxError("Missing operands for binary operator")
        operand1 = Icfp.parse_token(rest.pop(0), rest)
        operand2 = Icfp.parse_token(rest.pop(0), rest)
        return IcfpBinaryOp(tok, operator, operand1, operand2)


class IcfpConditional(Icfp):

    def __init__(self, tok, condition, true_clause, false_clause):
        Icfp.__init__(self, tok)
        self.condition = condition
        self.true_clause = true_clause
        self.false_clause = false_clause

    def _dbg_out(self, depth, level_tracker, seq_out, tree_out):
        super()._dbg_out(depth, level_tracker, seq_out, tree_out)
        level_tracker[depth] = True
        self.condition._dbg_out(depth + 1, level_tracker, seq_out, tree_out)
        self.true_clause._dbg_out(depth + 1, level_tracker, seq_out, tree_out)
        level_tracker.pop(depth, None)
        self.false_clause._dbg_out(depth + 1, level_tracker, seq_out, tree_out)

    def substitute(self, var_id, replacement):
        new_condition = self.condition.substitute(var_id, replacement)
        new_true_clause = self.true_clause.substitute(var_id, replacement)
        new_false_clause = self.false_clause.substitute(var_id, replacement)
        if (new_condition is self.condition and
            new_true_clause is self.true_clause and
            new_false_clause is self.false_clause):
            return self
        return IcfpConditional(self.token(), new_condition, new_true_clause,
                               new_false_clause)

    def eval(self):
        cond = self.condition.eval()
        return self.true_clause.eval() if cond else self.false_clause.eval()

    @staticmethod
    def _parse_token(tok, rest):
        if len(tok) != 1:
            raise SyntaxError("Malformed conditional")
        if tok[0] != '?':
            raise SyntaxError("Invalid conditional")
        if len(rest) < 3:
            raise SyntaxError("Missing operands for conditional")
        condition = Icfp.parse_token(rest.pop(0), rest)
        true_clause = Icfp.parse_token(rest.pop(0), rest)
        false_clause = Icfp.parse_token(rest.pop(0), rest)
        return IcfpConditional(tok, condition, true_clause, false_clause)

class IcfpLambda(Icfp):

    def __init__(self, tok, var_num, expr):
        Icfp.__init__(self, tok)
        self.var_num = var_num
        self.expr = expr

    def _dbg_out(self, depth, level_tracker, seq_out, tree_out):
        super()._dbg_out(depth, level_tracker, seq_out, tree_out)
        self.expr._dbg_out(depth + 1, level_tracker, seq_out, tree_out)

    def ref_var_id(self):
        return self.var_num

    def ref_expr(self):
        return self.expr

    def substitute(self, var_id, replacement):
        if self.var_num == var_id:
            return self  # Shadowing the same variable.
        new_expr = self.expr.substitute(var_id, replacement)
        if new_expr is self.expr:
            return self
        return IcfpLambda(self.token(), var_id, new_expr)

    def eval(self):
        return self

    @staticmethod
    def _parse_token(tok, rest):
        if tok[0] != 'L':
            raise SyntaxError("Invalid lambda expression")
        body = tok[1:]
        if len(body) == 0:
            raise SyntaxError("Missing base-94 number in lambda")
        var_num = _str_to_num(body)
        if len(rest) < 1:
            raise SyntaxError("Missing operand for lambda")
        expr = Icfp.parse_token(rest.pop(0), rest)
        return IcfpLambda(tok, var_num, expr)

class IcfpVariable(Icfp):

    def __init__(self, tok, var_num):
        Icfp.__init__(self, tok)
        self.var_num = var_num

    def var_id(self):
        return self.var_num

    def substitute(self, var_id, replacement):
        if self.var_num == var_id:
            return replacement
        return self

    def eval(self):
        raise ValueError(f"Evaluating unsubstituted variable '{self.tok}'")

    @staticmethod
    def _parse_token(tok):
        if tok[0] != 'v':
            raise SyntaxError("Invalid variable")
        body = tok[1:]
        if len(body) == 0:
            raise SyntaxError("Missing base-94 number in variable")
        var_num = _str_to_num(body)
        return IcfpVariable(tok, var_num)


class IcfpCodec(object):

    def __init__(self):
        self.enc_key = {}
        for i in range(94):
            self.enc_key[DEC_KEY[i]] = chr(i + CHARSET_DELTA)

    def encode(self, msg):
        out = ['S']
        for c in list(msg):
            out += self.enc_key[c]
        return "".join(out)

    def decode(self, msg):
        expr = []
        tokens = msg.split(" ")
        while tokens:
            tok = tokens.pop(0)
            if len(tok) == 0:
                continue
            expr.append(Icfp.parse_token(tok, tokens))
        if len(expr) != 1:
            raise SyntaxError("Must have exactly one expression to evaluate")
        return expr[0]


class IcfpLangDecoderRepl(code.InteractiveConsole):

    def __init__(self):
        code.InteractiveConsole.__init__(self)
        self.codec = IcfpCodec()

    def runsource(self, source, filename="<input>", symbol="single"):
        src = source.strip()
        if len(src) == 0:
            return
        try:
            expr = self.codec.decode(src)
            if logging.getLogger().isEnabledFor(logging.INFO):
                logging.info(f"\n{Icfp.get_dbg_out(expr)}")
            val = expr.eval()
            val_str = "".join(val) if isinstance(val, list) else str(val)
            print(f"{val_str}")
        except SyntaxError as syntax_err:
            print(f"ERROR: {syntax_err}\n")


def main():
    arg_parser = argparse.ArgumentParser(
        prog="IcfgLangRepl",
        description="A simple REPL for decoding ICFP 2024 messages.")
    arg_parser.add_argument("-l", "--log",
                            choices=["debug", "info", "warning", "error"],
                            default="info")
    args = arg_parser.parse_args()

    log_level = getattr(logging, args.log.upper(), None)
    logging.basicConfig(format="%(asctime)s %(levelname)s: %(message)s",
                        datefmt="%Y-%m-%d %H:%M:%S", level=log_level)

    repl = IcfpLangDecoderRepl()
    repl.interact(banner="ICFP Lang Decoder.", exitmsg="Bye!")


if __name__ == "__main__":
    main()
