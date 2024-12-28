#!/usr/bin/env python3
import icfp_lang
import unittest


class TestIcfpLang(unittest.TestCase):

    def setUp(self):
        self.codec = icfp_lang.IcfpCodec()

    def eval(self, src):
        val = self.codec.decode(src).eval()
        return "".join(val) if isinstance(val, list) else str(val)

    def test_boolean(self):
        self.assertEqual(self.eval("T"), str(True))
        self.assertEqual(self.eval("F"), str(False))

    def test_integer(self):
        self.assertEqual(self.eval("I/6"), str(1337))

    def test_string(self):
        self.assertEqual(self.eval("SB%,,/}Q/2,$_"), "Hello World!")

    def test_unary_op(self):
        self.assertEqual(self.eval("U- I$"), str(-3))
        self.assertEqual(self.eval("U! T"), str(False))
        self.assertEqual(self.eval("U! F"), str(True))
        self.assertEqual(self.eval("U# S4%34"), str(15818151))
        self.assertEqual(self.eval("U$ I4%34"), "test")

    def test_binary_op(self):
        self.assertEqual(self.eval("B+ I# I$"), str(5))
        self.assertEqual(self.eval("B- I$ I#"), str(1))
        self.assertEqual(self.eval("B* I$ I#"), str(6))
        self.assertEqual(self.eval("B/ U- I( I#"), str(-3))
        self.assertEqual(self.eval("B% U- I( I#"), str(-1))
        self.assertEqual(self.eval("B< I$ I#"), str(False))
        self.assertEqual(self.eval("B> I$ I#"), str(True))
        self.assertEqual(self.eval("B= I$ I#"), str(False))
        self.assertEqual(self.eval("B| T F"), str(True))
        self.assertEqual(self.eval("B& T F"), str(False))
        self.assertEqual(self.eval("B. S4% S34"), "test")
        self.assertEqual(self.eval("BT I$ S4%34"), "tes")
        self.assertEqual(self.eval("BD I$ S4%34"), "t")

    def test_conditional(self):
        self.assertEqual(self.eval("? B> I# I$ S9%3 S./"), "no")

    def test_lambda(self):
        self.assertEqual(self.eval("B$ B$ L# L$ v# B. SB%,,/ S}Q/2,$_ IK"),
                         "Hello World!")
        self.assertEqual(self.eval("B$ L# B$ L\" B+ v\" v\" B* I$ I# v8"),
                         str(12))
        self.assertEqual(self.eval(
            "B$ B$ L\" B$ L# B$ v\" B$ v# v# L# B$ v\" B$ v# v# L\" L# ? B= "
            "v# I! I\" B$ L$ B+ B$ v\" v$ B$ v\" v$ B- v# I\" I%"), str(16))
        self.assertEqual(self.eval(
            "B. SF B$ B$ L\" B$ L\" B$ L# B$ v\" B$ v# v# L# B$ v\" B$ v# v# "
            "L$ L# ? B= v# I\" v\" B. v\" B$ v$ B- v# I\" Sl I#,"), "L......."
            "................................................................"
            "................................................................"
            "................................................................")


if __name__ == '__main__':
    unittest.main()
