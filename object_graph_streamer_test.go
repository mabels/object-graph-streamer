package ObjectGraphStreamer

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/mabels/object-graph-streamer/mocks"

	"github.com/btcsuite/btcutil/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ObjectGraphStreamSuite struct {
	suite.Suite
	mockedSvalFn *SvalFnMock
}

func (s *ObjectGraphStreamSuite) SetupTest() {
	s.mockedSvalFn = &SvalFnMock{}
}

// ####################
// ## ObjectGraphStreamer tests ##
// ####################

func (s *ObjectGraphStreamSuite) TestSortWithOutWithString() {
	strValue := "string"
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		return assert.Equal(s.T(), strValue, (sVal.Val.(JsonValType)).Val)
	}))
	ObjectGraphStreamer(strValue, s.mockedSvalFn.Execute)
}

func (s *ObjectGraphStreamSuite) TestSortWithOutWithDate() {
	t := time.Now()
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		return assert.Equal(s.T(), t, (sVal.Val.(JsonValType)).Val)
	}))
	ObjectGraphStreamer(t, s.mockedSvalFn.Execute)
}

func (s *ObjectGraphStreamSuite) TestSortWithOutWithNumber() {
	n := 78
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		return assert.Equal(s.T(), n, (sVal.Val.(JsonValType)).Val)
	}))
	ObjectGraphStreamer(n, s.mockedSvalFn.Execute)
}

func (s *ObjectGraphStreamSuite) TestSortWithOutWithBoolean() {
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		return assert.Equal(s.T(), true, (sVal.Val.(JsonValType)).Val)
	}))
	ObjectGraphStreamer(true, s.mockedSvalFn.Execute)
}

func (s *ObjectGraphStreamSuite) TestSortWithOutWithArrayOfEmpty() {
	var emptySlice []int
	funcCallIdx := 1
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		result := false
		if funcCallIdx == 1 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), ARRAY_START, sVal.OutState.String())
		} else if funcCallIdx == 2 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), ARRAY_END, sVal.OutState.String())
		}
		funcCallIdx++
		return result
	}))
	ObjectGraphStreamer(emptySlice, s.mockedSvalFn.Execute)
}

func (s *ObjectGraphStreamSuite) TestSortWithOutWithArrayOf_1_2() {
	ar := []int{1, 2}
	funcCallIdx := 1
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		result := false
		if funcCallIdx == 1 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), ARRAY_START, sVal.OutState.String())
		} else if funcCallIdx == 2 {
			result = assert.Equal(s.T(), ar[0], (sVal.Val.(JsonValType)).Val)
		} else if funcCallIdx == 3 {
			result = assert.Equal(s.T(), ar[1], (sVal.Val.(JsonValType)).Val)
		} else if funcCallIdx == 4 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), ARRAY_END, sVal.OutState.String())
		}
		funcCallIdx++
		return result
	}))
	ObjectGraphStreamer(ar, s.mockedSvalFn.Execute)
}

func (s *ObjectGraphStreamSuite) TestSortWithOutWithArrayOf_1_2_3_4() {
	ar := [][]int{{1, 2}, {3, 4}}
	funcCallIdx := 1
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		result := false
		if funcCallIdx == 1 || funcCallIdx == 2 || funcCallIdx == 6 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), ARRAY_START, sVal.OutState.String())
		} else if funcCallIdx == 3 {
			result = assert.Equal(s.T(), ar[0][0], (sVal.Val.(JsonValType)).Val)
		} else if funcCallIdx == 4 {
			result = assert.Equal(s.T(), ar[0][1], (sVal.Val.(JsonValType)).Val)
		} else if funcCallIdx == 5 || funcCallIdx == 9 || funcCallIdx == 10 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), ARRAY_END, sVal.OutState.String())
		} else if funcCallIdx == 7 {
			result = assert.Equal(s.T(), ar[1][0], (sVal.Val.(JsonValType)).Val)
		} else if funcCallIdx == 8 {
			result = assert.Equal(s.T(), ar[1][1], (sVal.Val.(JsonValType)).Val)
		}
		funcCallIdx++
		return result
	}))
	ObjectGraphStreamer(ar, s.mockedSvalFn.Execute)
}

func (s *ObjectGraphStreamSuite) TestSortWithOutWithObjOfEmptyObj() {
	var obj struct{}
	funcCallIdx := 1
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		result := false
		if funcCallIdx == 1 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), OBJECT_START, sVal.OutState.String())
		} else if funcCallIdx == 2 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), OBJECT_END, sVal.OutState.String())
		}
		funcCallIdx++
		return result
	}))
	ObjectGraphStreamer(obj, s.mockedSvalFn.Execute)
}

func (s *ObjectGraphStreamSuite) TestSortWithOutWithObjOfObj_Y_1_X_2() {
	funcCallIdx := 1
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		result := false
		if funcCallIdx == 1 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), OBJECT_START, sVal.OutState.String())
		} else if funcCallIdx == 2 {
			result = assert.Equal(s.T(), "x", sVal.Attribute)
		} else if funcCallIdx == 3 {
			result = assert.Equal(s.T(), 2, (sVal.Val.(JsonValType)).Val)
		} else if funcCallIdx == 4 {
			result = assert.Equal(s.T(), "y", sVal.Attribute)
		} else if funcCallIdx == 5 {
			result = assert.Equal(s.T(), 1, (sVal.Val.(JsonValType)).Val)
		} else if funcCallIdx == 6 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), OBJECT_END, sVal.OutState.String())
		}
		funcCallIdx++
		return result
	}))
	ObjectGraphStreamer(struct {
		Y int `json:"y"`
		X int `json:"x"`
	}{Y: 1, X: 2}, s.mockedSvalFn.Execute)
}

func (s *ObjectGraphStreamSuite) TestSortWithOutWithObjOfObj_Y_B_1_A_2() {
	funcCallIdx := 1
	s.mockedSvalFn.On("Execute", mock.MatchedBy(func(sVal SVal) bool {
		result := false
		if funcCallIdx == 1 || funcCallIdx == 3 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), OBJECT_START, sVal.OutState.String())
		} else if funcCallIdx == 2 {
			result = assert.Equal(s.T(), "y", sVal.Attribute)
		} else if funcCallIdx == 4 {
			result = assert.Equal(s.T(), "a", sVal.Attribute)
		} else if funcCallIdx == 5 {
			result = assert.Equal(s.T(), 2, (sVal.Val.(JsonValType)).Val)
		} else if funcCallIdx == 6 {
			result = assert.Equal(s.T(), "b", sVal.Attribute)
		} else if funcCallIdx == 7 {
			result = assert.Equal(s.T(), 1, (sVal.Val.(JsonValType)).Val)
		} else if funcCallIdx == 8 || funcCallIdx == 9 {
			result = assert.Equal(s.T(), nil, sVal.Val)
			result = result && assert.Equal(s.T(), OBJECT_END, sVal.OutState.String())
		}
		funcCallIdx++
		return result
	}))

	type Obj struct {
		B int `json:"b"`
		A int `json:"a"`
	}
	ObjectGraphStreamer(struct {
		Y Obj `json:"y"`
	}{Y: Obj{
		B: 1,
		A: 2,
	}}, s.mockedSvalFn.Execute)
}

// #########################
// ## JSONCollector tests ##
// #########################
func (s *ObjectGraphStreamSuite) TestJSONCollectorEmptyObj() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, nil)
	var obj struct{}
	ObjectGraphStreamer(obj, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "{}", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollectorEmptyArray() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, nil)
	var emptySlice []int
	ObjectGraphStreamer(emptySlice, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "[]", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollector_X_Y_1_Z_x_Y_Z() {
	type Obj struct {
		Y int    `json:"y"`
		Z string `json:"z"`
	}
	var emptySlice []int
	var emptypObj struct{}

	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, nil)
	ObjectGraphStreamer(struct {
		X Obj      `json:"x"`
		Y struct{} `json:"y"`
		Z []int    `json:"z"`
	}{
		X: Obj{
			Y: 1,
			Z: "x",
		},
		Y: emptypObj,
		Z: emptySlice,
	}, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "{\"x\":{\"y\":1,\"z\":\"x\"},\"y\":{},\"z\":[]}", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollectorArray_xx() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, nil)
	ObjectGraphStreamer([]string{"xx"}, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "[\"xx\"]", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollectorArray_1_2() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, nil)
	ObjectGraphStreamer([]interface{}{1, "2"}, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "[1,\"2\"]", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollector_1_2_A() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, nil)
	ObjectGraphStreamer([]interface{}{1, []string{"2", "A"}, "E"}, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "[1,[\"2\",\"A\"],\"E\"]", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollectorIndent2EmptyObj() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, NewJsonProps(2, ""))
	var obj struct{}
	ObjectGraphStreamer(obj, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "{}", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollectorIndent2ArrayEmpty() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, NewJsonProps(2, ""))
	var emptySlice []int
	ObjectGraphStreamer(emptySlice, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "[]", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollectorIndent2_X_Y_1_Z_x() {
	type Obj struct {
		Y int    `json:"y"`
		Z string `json:"z"`
	}
	var emptySlice []int
	var emptypObj struct{}

	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, NewJsonProps(2, ""))
	ObjectGraphStreamer(struct {
		X Obj      `json:"x"`
		Y struct{} `json:"y"`
		Z []int    `json:"z"`
	}{
		X: Obj{
			Y: 1,
			Z: "x",
		},
		Y: emptypObj,
		Z: emptySlice,
	}, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "{\n  \"x\": {\n    \"y\": 1,\n    \"z\": \"x\"\n  },\n  \"y\": {},\n  \"z\": []\n}", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollector_Indent2_xx() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, NewJsonProps(2, ""))
	ObjectGraphStreamer([]string{"xx"}, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "[\n  \"xx\"\n]", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollector_Indent2_array_1_2() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, NewJsonProps(2, ""))
	ObjectGraphStreamer([]interface{}{1, "2"}, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "[\n  1,\n  \"2\"\n]", out)
}

func (s *ObjectGraphStreamSuite) TestJSONCollector_1_date444() {
	out := ""
	col := NewJsonCollector(func(str string) {
		out += str
	}, nil)
	ObjectGraphStreamer([]interface{}{1, time.UnixMilli(444).UTC()}, func(prob SVal) {
		col.Append(prob)
	})
	assert.Equal(s.T(), "[1,\"1970-01-01T00:00:00.444Z\"]", out)
}

// #########################
// ## HashCollector tests ##
// #########################
func (s *ObjectGraphStreamSuite) TestHashCollector_date() {
	h := NewHashCollector()
	ObjectGraphStreamer(time.UnixMilli(444).UTC(), func(prob SVal) {
		h.Append(prob)
	})
	assert.Equal(s.T(), "DzYqv3YaniBJWwqrNBn4534oTe4nL14TqcfVCguf9Yyv", h.Digest())
}

func (s *ObjectGraphStreamSuite) TestHashCollector_X_1_Y2() {
	h := NewHashCollector()
	ObjectGraphStreamer(struct {
		X int
		Y int
	}{
		X: 1,
		Y: 2,
	}, func(prob SVal) {
		h.Append(prob)
	})
	assert.Equal(s.T(), "DkPs9C3fYabdDLFxqMh4ZoTNr3xD1xYGFvYnJioF7V6H", h.Digest())
}

func (s *ObjectGraphStreamSuite) TestHashCollector_1() {
	h := NewHashCollector()
	type Obj struct {
		Y int    `json:"y"`
		Z string `json:"z"`
	}
	var emptySlice []int
	var emptypObj struct{}
	ObjectGraphStreamer(struct {
		X Obj       `json:"x"`
		Y struct{}  `json:"y"`
		Z []int     `json:"z"`
		D time.Time `json:"d"`
	}{
		X: Obj{
			Y: 1,
			Z: "x",
		},
		Y: emptypObj,
		Z: emptySlice,
		D: time.UnixMilli(444).UTC(),
	}, func(prob SVal) {
		h.Append(prob)
	})
	assert.Equal(s.T(), "5PvJAWGkaKAHax6tsaKGfPYm6JfXxZs15wRTDpSKaZ2G", h.Digest())
}

func (s *ObjectGraphStreamSuite) TestHashCollector_2() {
	h := NewHashCollector()
	type Obj struct {
		Y int    `json:"y"`
		Z string `json:"z"`
	}
	var emptySlice []int
	var emptypObj struct{}
	ObjectGraphStreamer(struct {
		X    Obj       `json:"x"`
		Y    struct{}  `json:"y"`
		Z    []int     `json:"z"`
		Date time.Time `json:"date"`
	}{
		X: Obj{
			Y: 2,
			Z: "x",
		},
		Y:    emptypObj,
		Z:    emptySlice,
		Date: time.UnixMilli(444).UTC(),
	}, func(prob SVal) {
		h.Append(prob)
	})
	assert.Equal(s.T(), "ECVWfmcNaUGkgvPZe7CojrnRNULxNczKXU8PGns6UDvr", h.Digest())
}

func (s *ObjectGraphStreamSuite) TestHashCollector_3() {
	h := NewHashCollector()
	type Obj struct {
		X int    `json:"x"`
		Z string `json:"z"`
	}
	var emptySlice []int
	var emptypObj struct{}
	ObjectGraphStreamer(struct {
		X    Obj       `json:"x"`
		Y    struct{}  `json:"y"`
		Z    []int     `json:"z"`
		Date time.Time `json:"date"`
	}{
		X: Obj{
			X: 1,
			Z: "x",
		},
		Y:    emptypObj,
		Z:    emptySlice,
		Date: time.UnixMilli(444).UTC(),
	}, func(prob SVal) {
		h.Append(prob)
	})
	assert.Equal(s.T(), "EoYNGMtap1k9iEAGeVtHmJwpMjQLKWJmR27SG6aC9fSg", h.Digest())
}

func (s *ObjectGraphStreamSuite) TestHashCollector_4() {
	h1 := NewHashCollector()
	type Obj struct {
		X int    `json:"x"`
		Z string `json:"z"`
	}
	var emptySlice []int
	var emptypObj struct{}
	ObjectGraphStreamer(struct {
		X    Obj       `json:"x"`
		Y    struct{}  `json:"y"`
		Z    []int     `json:"z"`
		Date time.Time `json:"date"`
	}{
		X: Obj{
			X: 1,
			Z: "x",
		},
		Y:    emptypObj,
		Z:    emptySlice,
		Date: time.UnixMilli(444).UTC(),
	}, func(prob SVal) {
		h1.Append(prob)
	})

	h2 := NewHashCollector()
	ObjectGraphStreamer(struct {
		Date time.Time `json:"date"`
		X    Obj       `json:"x"`
		Y    struct{}  `json:"y"`
		Z    []int     `json:"z"`
	}{
		Date: time.UnixMilli(444).UTC(),
		X: Obj{
			X: 1,
			Z: "x",
		},
		Y: emptypObj,
		Z: emptySlice,
	}, func(prob SVal) {
		h2.Append(prob)
	})

	assert.Equal(s.T(), h1.Digest(), h2.Digest())
}

func (s *ObjectGraphStreamSuite) TestHashCollector_3_InternalUpdate() {
	hashCalculator := sha256.New()

	type Obj struct {
		R int    `json:"r"`
		Z string `json:"z"`
	}
	var emptySlice []int
	var emptypObj struct{}
	expectedArgs := []string{"date", "1970-01-01T00:00:00.444Z", "x", "r", "1", "z", "u", "y", "z"}
	mck := &mocks.Hash{}
	idx := 0
	mck.On("Write", mock.MatchedBy(func(p []byte) bool {
		hashCalculator.Write(p)
		idx++
		return expectedArgs[idx-1] == string(p)
	})).Return(1, nil)

	t := struct {
		X    Obj       `json:"x"`
		Y    struct{}  `json:"y"`
		Z    []int     `json:"z"`
		Date time.Time `json:"date"`
	}{
		X: Obj{
			R: 1,
			Z: "u",
		},
		Y:    emptypObj,
		Z:    emptySlice,
		Date: time.UnixMilli(444).UTC(),
	}
	collector := &HashCollector{mck}
	ObjectGraphStreamer(t, func(prob SVal) {
		collector.Append(prob)
	})

	var nilBytes []byte
	mck.On("Sum", nilBytes).Return([]byte{})
	collector.Digest()
	mck.AssertNumberOfCalls(s.T(), "Sum", 1)

	assert.Equal(s.T(), "CwEMjUHV6BpDS7AGBAYqjY6qMKE6xC8Z56H5T2ZuUuXe", base58.Encode(hashCalculator.Sum(nil)))
}

func (s *ObjectGraphStreamSuite) TestSimpleHash() {
	type Data struct {
		Name string `json:"name"`
		Date string `json:"date"`
	}

	type KindData struct {
		Kind string `json:"kind"`
		Data Data   `json:"data"`
	}

	hashC := NewHashCollector()
	ObjectGraphStreamer(KindData{
		Kind: "test",
		Data: Data{
			Name: "object",
			Date: "2021-05-20",
		},
	}, func(sval SVal) {
		hashC.Append(sval)
	})
	assert.Equal(s.T(), "5zWhdtvKuGob1FbW9vUGPQKobcLtYYr5wU8AxQRVraeB", hashC.Digest())
}

// ##########################
// ## SimpleEnvelope tests ##
// ##########################

func TestObjectGraphStreamerSuite(t *testing.T) {
	suite.Run(t, new(ObjectGraphStreamSuite))
}

type SvalFnMock struct {
	mock.Mock
}

// Execute provides a mock function with given fields: prob
func (_m *SvalFnMock) Execute(prob SVal) {
	_m.Called(prob)
}
