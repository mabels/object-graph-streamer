from datetime import datetime, timezone, tzinfo
import hashlib
import json
import unittest
import unittest.mock

from object_graph_streamer import HashCollector, JsonCollector, JsonProps, objectGraphStreamer, OutState


class Mockdatetime:
    def now(self):
        return datetime.fromtimestamp(1624140000)


mockdatetime = Mockdatetime()


def toSVals(calls: list[unittest.mock.call]):
    return list(map(lambda x: list(map(lambda x: x.to_dict(), x.args)), calls))


class ObjectGraphStreamerTest(unittest.TestCase):

    def test_simple_hash(self):
        hashC = HashCollector()
        objectGraphStreamer({
            'kind': "test",
            'data': {
                'name': "object",
                'date': "2021-05-20",
            },
        },
            lambda sval: hashC.append(sval))
        self.assertEqual(
            hashC.digest(), "5zWhdtvKuGob1FbW9vUGPQKobcLtYYr5wU8AxQRVraeB")

    def test_sort_with_out_with_string(self):
        fn = unittest.mock.Mock()
        objectGraphStreamer("string", fn)
        self.assertEqual(toSVals(fn.mock_calls), [
                         [{'val': {'val': "string", }, 'outState': 'V', 'paths': []}]])

    def test_sort_with_out_with_date(self):
        fn = unittest.mock.Mock()
        objectGraphStreamer(datetime.fromtimestamp(0.444, tz=timezone.utc), fn)
        self.assertEqual(toSVals(fn.mock_calls),
                         [[{'val': {'val': datetime.fromtimestamp(0.444, tz=timezone.utc)}, 'outState': 'V', 'paths': []}]])

    def test_sort_with_out_with_number(self):
        fn = unittest.mock.Mock()
        objectGraphStreamer(4711, fn)
        self.assertEqual(toSVals(fn.mock_calls), [
                         [{'val': {'val': 4711}, 'outState': 'V', 'paths': []}]])

    def test_sort_with_out_with_boolean(self):
        fn = unittest.mock.Mock()
        objectGraphStreamer(False, fn)
        self.assertEqual(toSVals(fn.mock_calls), [
                         [{'val': {'val': False}, 'outState': 'V', 'paths': []}]])

    def test_sort_with_out_with_array_of_empty(self):
        fn = unittest.mock.Mock()
        objectGraphStreamer([], fn)
        self.assertEqual(toSVals(fn.mock_calls),
                         [[{'outState': "[", 'paths': ['[']}], [{'outState': "]", 'paths': [']']}]])

    def test_sort_with_out_with_array_of_1_2(self):
        fn = unittest.mock.Mock()
        objectGraphStreamer([1, 2], fn)
        self.assertEqual(toSVals(fn.mock_calls), [
            [{'outState': "[", 'paths': ['[']}],
            [{'val': {'val': 1}, 'outState': 'V', 'paths': ['[', '0']}],
            [{'val': {'val': 2}, 'outState': 'V', 'paths': ['[', '1']}],
            [{'outState': "]", 'paths': [']']}],
        ])

    def test_sort_with_out_with_array_of_1_2_3_4(self):
        fn = unittest.mock.Mock()
        objectGraphStreamer(
            [
                [1, 2],
                [3, 4],
            ],
            fn
        )
        self.assertEqual(toSVals(fn.mock_calls), [
            [{'outState': "[", 'paths': ['[']}],
            [{'outState': "[", 'paths': ['[', '0', '[']}],
            [{'val': {'val': 1}, 'outState': 'V','paths': ['[', '0', '[', '0']}],
            [{'val': {'val': 2}, 'outState': 'V','paths': ['[', '0', '[', '1']}],
            [{'outState': "]", 'paths': ['[', '0', ']']}],
            [{'outState': "[", 'paths': ['[', '1', '[']}],
            [{'val': {'val': 3,}, 'outState': 'V', 'paths': ['[', '1', '[', '0']}],
            [{'val': {'val': 4}, 'outState': 'V', 'paths': ['[', '1', '[', '1']}],
            [{'outState': "]", 'paths': ['[', '1', ']']}],
            [{'outState': "]", 'paths': [']']}],
        ])

    def test_sort_with_out_with_obj_of_empty_obj(self):
        fn = unittest.mock.Mock()
        objectGraphStreamer({}, fn)
        self.assertEqual(toSVals(fn.mock_calls),
                         [[{'outState': "{",  'paths': ['{']}], [{'outState': "}", 'paths': ['}']}]])

    def test_sort_with_out_with_obj_of_obj_y_1_x_2(self):
        fn = unittest.mock.Mock()
        objectGraphStreamer({'y': 1, 'x': 2}, fn)
        self.assertEqual(toSVals(fn.mock_calls), [
            [{'outState': "{", 'paths': ['{']}],
            [{'attribute': "x", 'outState': 'A', 'paths': ['{', 'x']}],
            [{'val': {'val': 2}, 'outState': 'V', 'paths': ['{', 'x']}],
            [{'attribute': "y", 'outState': 'A', 'paths': ['{', 'y']}],
            [{'val': {'val': 1}, 'outState': 'V', 'paths': ['{', 'y']}],
            [{'outState': "}", 'paths': ['}']}],
        ])

    def test_sort_with_out_with_obj_of_obj_y_b_1_a_2(self):
        self.maxDiff = None
        fn = unittest.mock.Mock()
        objectGraphStreamer({'y': {'b': 1, 'a': 2}}, fn)
        self.assertEqual(toSVals(fn.mock_calls), [
            [{'outState': "{", 'paths': ['{']}],
            [{'attribute': "y", 'outState': 'A', 'paths': ['{', 'y']}],
            [{'outState': "{", 'paths': ['{', 'y', '{']}],
            [{'attribute': "a", 'outState': 'A', 'paths': ['{', 'y', '{', 'a']}],
            [{'val': {'val': 2}, 'outState': 'V', 'paths': ['{', 'y', '{', 'a']}],
            [{'attribute': "b", 'outState': 'A', 'paths': ['{', 'y', '{', 'b']}],
            [{'val': {'val': 1}, 'outState': 'V', 'paths': ['{', 'y', '{', 'b']}],
            [{'outState': "}", 'paths': ['{', 'y', '}']}],
            [{'outState': "}", 'paths': ['}']}],
        ])

    def test_JSONCollector_empty_obj(self):
        out = []
        json = JsonCollector(lambda o: out.append(o))
        objectGraphStreamer({}, lambda o: json.append(o))
        self.assertEqual("".join(out), "{}")

    def test_JSONCollector_empty_array(self):
        out = []
        json = JsonCollector(lambda o: out.append(o))
        objectGraphStreamer([], lambda o: json.append(o))
        self.assertEqual("".join(out), "[]")

    def test_JSONCollector_x_y_1_z_x_y_z(self):
        out = []
        json = JsonCollector(lambda o: out.append(o))
        objectGraphStreamer({'x': {'y': 1, 'z': "x"}, 'y': {}, 'z': []},
                            lambda o: json.append(o))
        self.assertEqual("".join(out), '{"x":{"y":1,"z":"x"},"y":{},"z":[]}')

    def test_JSONCollector_array_xx(self):
        out = []
        json = JsonCollector(lambda o: out.append(o))
        objectGraphStreamer(["xx"], lambda o: json.append(o))
        self.assertEqual("".join(out), '["xx"]')

    def test_JSONCollector_array_1_2(self):
        out = []
        json = JsonCollector(lambda o: out.append(o))
        objectGraphStreamer([1, "2"], lambda o: json.append(o))
        self.assertEqual("".join(out), '[1,"2"]')

    def test_JSONCollector_1_2_A(self):
        out = []
        json = JsonCollector(lambda o: out.append(o))
        objectGraphStreamer([1, ["2", "A"], "E"], lambda o: json.append(o))
        self.assertEqual("".join(out), '[1,["2","A"],"E"]')

    def test_JSONCollector_indent_2_empty_obj(self):
        out = []
        json = JsonCollector(lambda o: out.append(o),
                             JsonProps(**{'indent': 2}))
        objectGraphStreamer({}, lambda o: json.append(o))
        self.assertEqual("".join(out), "{}")

    def test_JSONCollector_indent_2_array_empty(self):
        out = []
        json = JsonCollector(lambda o: out.append(o),
                             JsonProps(**{'indent': 2}))
        objectGraphStreamer([], lambda o: json.append(o))
        self.assertEqual("".join(out), "[]")

    def test_JSONCollector_indent_2_x_y_1_z_x(self):
        out = []
        json = JsonCollector(lambda o: out.append(o),
                             JsonProps(**{'indent': 2}))
        objectGraphStreamer({'x': {'y': 1, 'z': "x"}, 'y': {}, 'z': []},
                            lambda o: json.append(o))
        self.assertEqual(
            "".join(out), '{\n  "x": {\n    "y": 1,\n    "z": "x"\n  },\n  "y": {},\n  "z": []\n}')

    def test_JSONCollector_indent_2_xx(self):
        out = []
        json = JsonCollector(lambda o: out.append(o),
                             JsonProps(**{'indent': 2}))
        objectGraphStreamer(["xx"], lambda o: json.append(o))
        self.assertEqual("".join(out), '[\n  "xx"\n]')

    def test_JSONCollector_indent_2_array_1_2(self):
        out = []
        json = JsonCollector(lambda o: out.append(o),
                             JsonProps(**{'indent': 2}))
        objectGraphStreamer([1, "2"], lambda o: json.append(o))
        self.assertEqual("".join(out), '[\n  1,\n  "2"\n]')

    def test_JSONCollector_1_date444(self):
        out = []
        json = JsonCollector(lambda o: out.append(o))
        obj = [1, datetime.fromtimestamp(0.444, tz=timezone.utc)]
        objectGraphStreamer(obj, lambda o: json.append(o))
        self.assertEqual("[1,\"1970-01-01T00:00:00.444Z\"]", "".join(out))

    def test_HashCollector_1(self):
        hash = HashCollector()
        objectGraphStreamer({'x': {'y': 1, 'z': "x"}, 'y': {}, 'z': [],
                             'd': datetime.fromtimestamp(0.444, tz=timezone.utc)}, lambda o: hash.append(o))
        self.assertEqual(
            hash.digest(), "5PvJAWGkaKAHax6tsaKGfPYm6JfXxZs15wRTDpSKaZ2G")

    def test_HashCollector_2(self):
        hash = HashCollector()
        objectGraphStreamer({'x': {'y': 2, 'z': "x"}, 'y': {}, 'z': [],
                             'date': datetime.fromtimestamp(0.444, tz=timezone.utc)}, lambda o: hash.append(o))
        self.assertEqual(
            hash.digest(), "ECVWfmcNaUGkgvPZe7CojrnRNULxNczKXU8PGns6UDvr")

    def test_HashCollector_3(self):
        hash = HashCollector()
        objectGraphStreamer({'x': {'x': 1, 'z': "x"}, 'y': {}, 'z': [],
                             'date': datetime.fromtimestamp(0.444, tz=timezone.utc)}, lambda o: hash.append(o))
        self.assertEqual(
            hash.digest(), "EoYNGMtap1k9iEAGeVtHmJwpMjQLKWJmR27SG6aC9fSg")

    def test_HashCollector_4(self):
        hash1 = HashCollector()
        objectGraphStreamer({'x': {'x': 1, 'z': "x"}, 'y': {}, 'z': [],
                             'date': datetime.fromtimestamp(0.444, tz=timezone.utc)}, lambda o: hash1.append(o))
        hash2 = HashCollector()
        objectGraphStreamer({'date': datetime.fromtimestamp(0.444, tz=timezone.utc), 'x': {
            'x': 1, 'z': "x"}, 'y': {}, 'z': []}, lambda o: hash2.append(o))
        self.assertEqual(hash1.digest(), hash2.digest())

    def test_HashCollector_3_internal_update(self):
        class MockHash:
            hash: any
            mockUpdate: any

            def __init__(self) -> None:
                self.mockUpdate = unittest.mock.Mock()
                self.hash = hashlib.new('sha256')

            def update(self, a: any):
                self.mockUpdate(a)
                self.hash.update(a)

            def digest(self):
                return self.hash.digest()

        hashCollector = HashCollector(MockHash())
        objectGraphStreamer({'x': {'r': 1, 'z': "u"}, 'y': {}, 'z': [], 'date': datetime.fromtimestamp(
            0.444, tz=timezone.utc)}, lambda o: hashCollector.append(o))
        result = list(map(lambda m: m.args[0].decode(
        ), hashCollector.hash.mockUpdate.mock_calls))
        self.assertEqual(result,
                         ["date", "1970-01-01T00:00:00.444Z", "x", "r", "1", "z", "u", "y", "z"])
        self.assertEqual(hashCollector.digest(),
                         "CwEMjUHV6BpDS7AGBAYqjY6qMKE6xC8Z56H5T2ZuUuXe")


if __name__ == '__main__':
    unittest.main()
