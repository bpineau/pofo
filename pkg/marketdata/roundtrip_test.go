package marketdata

import (
	"math"
	"testing"
)

// fixture is a window of REAL consecutive closes around a known event, with the
// index of the session under scrutiny. Dates do not matter to dropRoundTrips,
// only the order, so the fixtures carry closes alone. Every one of them was
// read out of the live provider series (2026-08) by the plausibility campaign
// that motivated this pass.
type fixture struct {
	at     int
	closes []float64
}

// series builds Points from a fixture, one calendar day apart.
func (f fixture) series() []Point { return pts(f.closes...) }

// legs returns the two returns straddling the event, in percent.
func (f fixture) legs() (float64, float64) {
	c := f.closes
	return (c[f.at]/c[f.at-1] - 1) * 100, (c[f.at+1]/c[f.at] - 1) * 100
}

var (
	// IE00B5M4WH52 around 2015-07-01 (61 sessions, the event at index 30).
	iemlLocalGovt2015 = fixture{at: 30, closes: []float64{
		39.4808, 39.3446, 39.5189, 39.3228, 38.5821, 38.5113,
		38.3915, 38.6093, 38.2826, 38.3289, 38.2145, 38.0021,
		37.4956, 37.4602, 37.6508, 37.9531, 38.0810, 37.8714,
		37.8469, 37.6372, 37.8305, 38.2826, 38.0075, 38.7074,
		38.0484, 38.1954, 38.2580, 38.3338, 37.8089, 37.8397,
		46.1382, 38.0250, 37.9941, 37.8426, 37.4046, 37.4664,
		37.6236, 37.8285, 37.8201, 38.0952, 37.8678, 37.9296,
		37.8454, 37.6039, 37.7780, 37.5422, 37.4383, 37.1351,
		36.9218, 36.9667, 37.1014, 36.7702, 36.8881, 36.6101,
		36.6354, 36.3238, 36.1048, 36.1806, 36.2648, 35.9504,
		36.0908,
	}}
	// IE00B2NPKV68 around 2017-06-27 (61 sessions, the event at index 30).
	iembHardCurrency2017 = fixture{at: 30, closes: []float64{
		76.2199, 76.3544, 76.1997, 75.9037, 76.1661, 76.1997,
		76.2131, 76.3477, 76.5226, 76.5293, 76.6033, 76.5226,
		76.6840, 77.0339, 77.0204, 77.1752, 77.2088, 76.9330,
		76.9868, 77.0002, 76.9195, 77.3769, 76.9804, 76.9668,
		76.9871, 76.6020, 76.5344, 76.5885, 76.8115, 76.9196,
		68.9810, 76.5479, 76.4669, 76.2440, 76.0885, 76.1493,
		75.8521, 75.5751, 75.4197, 75.9466, 75.7980, 76.2912,
		76.1624, 76.4132, 76.5081, 76.6233, 76.8199, 76.8674,
		76.9148, 76.7928, 76.5623, 76.8131, 76.8877, 76.8605,
		77.0978, 77.0368, 77.0843, 77.3215, 77.2198, 77.3079,
		77.4096,
	}}
	// BTAL around 2015-04-29 (61 sessions, the event at index 30).
	btalAntiBeta2015 = fixture{at: 30, closes: []float64{
		17.5644, 17.5644, 17.5644, 17.5644, 17.5644, 17.5644,
		17.8246, 17.8593, 17.3909, 17.3909, 17.4343, 17.4343,
		17.4343, 17.4343, 17.4343, 17.2695, 17.1481, 17.0093,
		17.0093, 17.0093, 17.0093, 17.0873, 17.0873, 16.7404,
		16.7664, 16.7664, 16.7664, 17.0180, 16.8271, 16.7317,
		11.2412, 16.7317, 16.4281, 16.4281, 16.3154, 16.4281,
		16.6537, 16.7230, 17.2348, 17.2348, 17.2348, 16.6883,
		16.6883, 16.5756, 16.5756, 16.5756, 16.8705, 16.6537,
		16.6537, 16.6537, 16.6537, 16.6537, 16.7057, 16.7057,
		16.7057, 16.7057, 16.7057, 16.7057, 16.7057, 16.6190,
		16.6190,
	}}
	// IE00B8GF1M35 around 2012-10-29 (61 sessions, the event at index 3).
	globalRealEstate2012 = fixture{at: 3, closes: []float64{
		20.9772, 20.9863, 20.8848, 16.1673, 20.9694, 20.9129,
		21.2028, 21.3160, 21.0967, 21.1533, 21.0048, 21.0048,
		20.8704, 20.7714, 20.8280, 20.6865, 20.3895, 20.3046,
		20.7290, 20.7502, 20.7360, 20.8563, 20.9765, 20.9553,
		21.0260, 20.9199, 21.0331, 21.1250, 21.2382, 21.2170,
		21.1816, 21.2806, 21.4221, 21.5140, 21.5564, 21.4786,
		21.3796, 21.4008, 21.4857, 21.6484, 21.6943, 21.8464,
		21.8712, 21.8712, 21.7615, 21.7757, 21.7757, 22.1930,
		22.1505, 22.0692, 22.1152, 22.0197, 22.1717, 22.1965,
		22.1505, 22.2177, 22.2531, 22.2884, 22.2884, 22.2672,
		22.2672,
	}}
	// VFINX around 1987-10-19 (61 sessions, the event at index 30).
	sp500Crash1987 = fixture{at: 30, closes: []float64{
		13.6187, 13.4822, 13.5035, 13.6400, 13.8490, 13.8959,
		13.6699, 13.5462, 13.5504, 13.5462, 13.3628, 13.7466,
		13.8191, 13.7637, 13.7850, 13.9952, 13.8536, 13.8579,
		14.0981, 14.1281, 14.1324, 13.7550, 13.7250, 13.5363,
		13.4119, 13.3261, 13.5491, 13.1502, 12.8414, 12.1766,
		9.6847, 10.1994, 11.1258, 10.6926, 10.6926, 9.8091,
		10.0492, 10.0492, 10.5425, 10.8470, 11.0229, 10.8127,
		10.7355, 10.9886, 10.8127, 10.5039, 10.3237, 10.4481,
		10.7355, 10.6111, 10.6626, 10.5039, 10.6154, 10.3752,
		10.4653, 10.5039, 10.6540, 10.5554, 10.5554, 9.9592,
		10.0407,
	}}
	// VFINX around 2020-03-12 (61 sessions, the event at index 30).
	sp500Covid2020 = fixture{at: 30, closes: []float64{
		276.8884, 277.7951, 272.8959, 274.8738, 278.9946, 282.1356,
		283.1246, 281.6502, 283.7564, 284.2417, 286.0824, 285.7161,
		286.2839, 285.4596, 286.8516, 285.7618, 282.7675, 273.3171,
		265.0481, 264.0407, 252.4293, 250.4057, 261.9255, 254.5722,
		265.3228, 256.3854, 252.0173, 232.8220, 244.3243, 232.4083,
		210.3508, 229.9533, 202.4068, 214.5067, 203.4090, 204.3652,
		195.5385, 189.8104, 207.6476, 210.0474, 223.1586, 215.6468,
		222.9012, 219.3522, 209.6704, 214.4975, 211.2794, 226.1285,
		225.7699, 233.5208, 236.9320, 234.5414, 241.7406, 236.4354,
		237.8055, 244.1864, 239.8190, 232.4726, 237.8055, 237.6859,
		241.0050,
	}}
	// SPY around 2008-10-13 (61 sessions, the event at index 30).
	sp500October2008 = fixture{at: 30, closes: []float64{
		92.6488, 92.0733, 91.9942, 89.2246, 89.5052, 91.3540,
		88.6419, 89.0016, 90.2893, 90.7065, 86.3903, 87.8362,
		83.8868, 86.3758, 89.8062, 87.7730, 85.7761, 86.0510,
		87.3968, 87.4402, 80.5882, 83.9238, 83.9744, 80.9283,
		79.8357, 75.7694, 72.3760, 70.5527, 65.6254, 64.0336,
		73.3311, 72.2458, 65.1334, 67.8467, 67.4414, 71.4933,
		69.3588, 65.5820, 66.3417, 62.9772, 60.7414, 67.8394,
		67.3474, 69.6772, 70.0607, 70.2633, 72.6510, 69.5976,
		65.7411, 67.9118, 67.0218, 64.9525, 62.0944, 65.9654,
		62.6733, 61.8412, 63.0061, 58.9688, 54.5913, 57.5361,
		61.5229,
	}}
	// IE00BGCSB447 around 2020-03-19 (61 sessions, the event at index 30).
	ernaDislocation2020 = fixture{at: 30, closes: []float64{
		5.2305, 5.2310, 5.2320, 5.2325, 5.2330, 5.2335,
		5.2350, 5.2345, 5.2345, 5.2340, 5.2340, 5.2350,
		5.2390, 5.2370, 5.2360, 5.2370, 5.2320, 5.2360,
		5.2395, 5.2415, 5.2445, 5.2410, 5.2205, 5.2420,
		5.2265, 5.1980, 5.1950, 5.1820, 5.1260, 5.0443,
		4.8810, 4.7920, 4.8800, 4.9437, 5.0000, 5.0390,
		5.0470, 5.0850, 5.0510, 5.1185, 5.1600, 5.1500,
		5.1625, 5.1545, 5.1630, 5.1830, 5.2300, 5.2085,
		5.1970, 5.2185, 5.2200, 5.2200, 5.2065, 5.2125,
		5.2170, 5.2120, 5.2195, 5.2235, 5.2300, 5.2315,
		5.2315,
	}}
	// ^990100-USD-STRD around 2001-09-24 (61 sessions, the event at index 30).
	worldShape2001 = fixture{at: 30, closes: []float64{
		1069.7000, 1070.4000, 1055.6000, 1048.9000, 1053.9000, 1063.1000,
		1063.1000, 1059.1000, 1045.1000, 1045.1000, 1042.3000, 1040.5000,
		1042.9000, 1057.7000, 1056.2000, 1043.3000, 1032.3000, 1017.6000,
		1018.2000, 1010.6000, 1011.2000, 1004.5000, 987.4000, 972.2000,
		968.1000, 962.8000, 968.1000, 906.9000, 879.5000, 855.9000,
		968.1000, 898.9000, 899.2000, 905.8000, 925.7000, 922.2000,
		932.9000, 941.1000, 952.5000, 952.5000, 948.7000, 939.8000,
		957.1000, 969.3000, 969.8000, 961.5000, 967.9000, 964.0000,
		952.9000, 949.4000, 958.9000, 965.9000, 967.6000, 968.8000,
		977.1000, 959.6000, 941.5000, 943.1000, 958.2000, 960.4000,
		973.5000,
	}}
	// FR0000288946 around 1989-05-17 (61 sessions, the event at index 30).
	axaCourtTerme1989 = fixture{at: 30, closes: []float64{
		1102.2252, 1101.9265, 1104.1154, 1105.0354, 1105.7655, 1104.9354,
		1106.0323, 1105.3255, 1106.2255, 1106.7756, 1107.1456, 1107.4243,
		1110.3459, 1109.9559, 1114.3762, 1116.2864, 1112.7026, 1113.2361,
		1113.3762, 1113.0561, 1112.6261, 1115.7945, 1115.5263, 1116.6164,
		1117.4396, 1117.2565, 1115.7964, 1115.9664, 1115.4774, 1117.5165,
		1034.6294, 1034.2593, 1035.5765, 1035.2694, 1037.0796, 1036.3795,
		1036.6095, 1035.2594, 1033.9293, 1034.0993, 1034.2516, 1035.3294,
		1036.3295, 1040.2198, 1038.1297, 1039.1597, 1039.4598, 1039.9498,
		1040.4599, 1041.3499, 1045.3203, 1045.6703, 1046.3251, 1045.4403,
		1046.4504, 1046.4504, 1046.4104, 1046.4638, 1047.1004, 1047.6105,
		1047.3404,
	}}
)

// dropped reports whether dropRoundTrips removed the fixture's event session,
// judged by the class band the real asset carries.
func dropped(f fixture, class string, leverage float64) bool {
	band, ok := ClassBand(class)
	if !ok {
		panic("unknown asset class " + class)
	}
	in := f.series()
	out := dropRoundTrips(in, band.Scale(leverage))
	if len(out) == len(in) {
		return false
	}
	for _, p := range out {
		if p.Close == f.closes[f.at] {
			return false // some other point went instead
		}
	}
	return len(out) == len(in)-1
}

// TestDropRoundTripsKills covers the four impossible one-session round trips
// the plausibility campaign found in the bundled catalog. Each is confirmed by
// a sibling listing of the same fund showing no such move.
func TestDropRoundTripsKills(t *testing.T) {
	for _, tc := range []struct {
		name     string
		f        fixture
		class    string
		leverage float64
	}{
		{"IE00B5M4WH52 2015-07-01", iemlLocalGovt2015, "government-bond", 1},
		{"IE00B2NPKV68 2017-06-27", iembHardCurrency2017, "government-bond", 1},
		{"IE00B8GF1M35 2012-10-29", globalRealEstate2012, "real-estate", 1},
		{"BTAL 2015-04-29", btalAntiBeta2015, "other", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.f.series()
			out := dropRoundTrips(in, mustBand(t, tc.class).Scale(tc.leverage))
			r1, r2 := tc.f.legs()
			if !dropped(tc.f, tc.class, tc.leverage) {
				t.Fatalf("legs %+.1f %% / %+.1f %% survived: %d points in, %d out", r1, r2, len(in), len(out))
			}
			// The neighbours now meet, carrying the true two-session move.
			net := (1+r1/100)*(1+r2/100) - 1
			got := out[tc.f.at].Close/out[tc.f.at-1].Close - 1
			if math.Abs(got-net) > 1e-9 {
				t.Fatalf("junction return %+.4f %%, want the round trip's %+.4f %%", got*100, net*100)
			}
		})
	}
}

// TestDropRoundTripsSpares is the other half of the contract, and the more
// important one: real history, however violent, must come through whole. Each
// case names the test that saves it, so a future loosening has to argue with a
// specific line.
func TestDropRoundTripsSpares(t *testing.T) {
	for _, tc := range []struct {
		name     string
		f        fixture
		class    string
		leverage float64
		why      string
	}{
		{"1987-10-19 and its rebound", sp500Crash1987, "equity", 1,
			"-20.5 % then +5.3 %: the legs leave 16 % standing, no round trip"},
		{"2020-03-12/13/16", sp500Covid2020, "equity", 1,
			"-9.5 % then +9.3 % against a ~4 % local sigma: two sigmas, and below the equity leg floor"},
		{"October 2008", sp500October2008, "equity", 1,
			"+14.5 % then -1.5 %: nothing reverses"},
		{"ERNA 2020-03-19", ernaDislocation2020, "money-market", 1,
			"-3.2 % then -1.8 %: same direction, a real ultrashort credit dislocation"},
		{"MSCI World 2001-09-24", worldShape2001, "other", 1,
			"+13.1 % then -7.1 %: 5 % left standing, so this pass declines it (simgen's shape despike takes it)"},
		{"AXA Court Terme 1989-05-17", axaCourtTerme1989, "money-market", 1,
			"-7.4 % and no return: an annual distribution in a price NAV, a level change and not a spike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.f.series()
			out := dropRoundTrips(in, mustBand(t, tc.class).Scale(tc.leverage))
			if len(out) != len(in) {
				r1, r2 := tc.f.legs()
				t.Fatalf("legs %+.1f %% / %+.1f %% were cleaned away (%s)", r1, r2, tc.why)
			}
		})
	}
}

func mustBand(t *testing.T, class string) Band {
	t.Helper()
	b, ok := ClassBand(class)
	if !ok {
		t.Fatalf("unknown asset class %q", class)
	}
	return b
}

// TestDropRoundTripsGuards covers the conditions that are not about a specific
// asset: too short a series to estimate anything, a rate-symbol exemption
// enforced upstream, and the class band widening with leverage.
func TestDropRoundTripsGuards(t *testing.T) {
	band := mustBand(t, "government-bond")

	t.Run("too short to judge", func(t *testing.T) {
		short := iemlLocalGovt2015
		short.closes = short.closes[15:45] // 30 points, below 2*roundTripWindow
		if got := dropRoundTrips(pts(short.closes...), band); len(got) != 30 {
			t.Fatalf("a 30-point series must be left alone, got %d points", len(got))
		}
	})

	t.Run("rate symbols never reach the pass", func(t *testing.T) {
		// cleanQuotes is the gate: a yield level crossing zero produces ratios
		// that mean nothing, so no cleaner runs on it.
		in := iemlLocalGovt2015.series()
		if got := cleanQuotes("^TNX", in); len(got) != len(in) {
			t.Fatalf("^TNX was cleaned: %d points in, %d out", len(in), len(got))
		}
	})

	t.Run("leverage widens the leg floor", func(t *testing.T) {
		// The same round trip, judged as a 3x fund of the same class: three
		// times the volatility ceiling, three times the leg it takes to
		// surprise, and -10.3 % no longer does.
		if !dropped(iembHardCurrency2017, "government-bond", 1) {
			t.Fatal("the unlevered case must be cleaned")
		}
		if dropped(iembHardCurrency2017, "government-bond", 3) {
			t.Fatal("at leverage 3 the leg is ordinary volatility and must survive")
		}
	})
}

// TestSpikeLegOrdersTheClasses pins the relationship the whole design rests on:
// the leg it takes to suspect a print scales with what the class can do, and
// stays far below the move the doctor calls implausible on its own.
func TestSpikeLegOrdersTheClasses(t *testing.T) {
	for _, class := range []string{"money-market", "aggregate-bond", "government-bond", "equity", "other"} {
		b := mustBand(t, class)
		if b.SpikeLeg() >= b.Move {
			t.Errorf("%s: SpikeLeg %.1f %% must stay below Move %.1f %%", class, b.SpikeLeg()*100, b.Move*100)
		}
	}
	if mustBand(t, "money-market").SpikeLeg() >= mustBand(t, "equity").SpikeLeg() {
		t.Error("a money-market leg floor must sit below an equity one")
	}
}
