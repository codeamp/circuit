package gota

import "testing"

func TestCMO(t *testing.T) {
	list := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	expList := []float64{100, 100, 100, 100, 100, 80, 60, 40, 20, 0, -20, -40, -60, -80, -100, -100, -100, -100, -100}

	cmo := NewCMO(10)
	var actList []float64
	for _, v := range list {
		if vOut := cmo.Add(v); cmo.Warmed() {
			actList = append(actList, vOut)
		}
	}

	if diff := diffFloats(expList, actList, 1E-7); diff != "" {
		t.Errorf("unexpected floats:\n%s", diff)
	}
}

func TestCMOS(t *testing.T) {
	list := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	// expList is generated by the following code:
	// expList, _ := talib.Cmo(list, 10, nil)
	expList := []float64{100, 100, 100, 100, 100, 80, 61.999999999999986, 45.79999999999999, 31.22, 18.097999999999992, 6.288199999999988, -4.340620000000012, -13.906558000000008, -22.515902200000014, -30.264311980000013, -37.23788078200001, -43.51409270380002, -49.16268343342002, -54.24641509007802}

	cmo := NewCMOS(10, WarmSMA)
	var actList []float64
	for _, v := range list {
		if vOut := cmo.Add(v); cmo.Warmed() {
			actList = append(actList, vOut)
		}
	}

	if diff := diffFloats(expList, actList, 1E-7); diff != "" {
		t.Errorf("unexpected floats:\n%s", diff)
	}
}
