package firebook

import (
	"reflect"
	"strings"
	"testing"
)

// The recharge staircase carries no frozen array: every step is recomputed at
// render time from four dated constants. These tests pin those constants to
// the law as verified on 2026-07-30 (service-public.gouv.fr, art. 779 I, 784
// and 790 G CGI), then check that the arithmetic the plate draws is the
// arithmetic the article states in prose.

func TestRechargeConstantsMatchTheLaw(t *testing.T) {
	if succAbattementDirect != 100000 {
		t.Errorf("art. 779 I CGI: allowance is 100 000 EUR, the plate uses %.0f", succAbattementDirect)
	}
	if succDonFamilial != 31865 {
		t.Errorf("art. 790 G CGI: family cash gift is 31 865 EUR, the plate uses %.0f", succDonFamilial)
	}
	if succDonAgeMax != 80 {
		t.Errorf("art. 790 G CGI: the donor must be under 80, the plate uses %d", succDonAgeMax)
	}
	if succRappelAns != 15 {
		t.Errorf("art. 784 CGI: the recall runs over 15 years, the plate uses %d", succRappelAns)
	}
}

// One wave, for the article's household: a couple and two adult children.
func TestRechargeStepSize(t *testing.T) {
	for _, tc := range []struct {
		age  int
		want float64
	}{
		{55, 527460}, // 2 x 2 x (100 000 + 31 865)
		{70, 527460},
		{79, 527460},
		{80, 400000}, // art. 790 G is over: the donor is no longer under 80
		{85, 400000},
		{90, 400000},
	} {
		if got := succStep(tc.age); got != tc.want {
			t.Errorf("wave at %d: %.0f EUR, expected %.0f", tc.age, got, tc.want)
		}
	}

	// The article's own two figures, checked against the same constants.
	// "Un couple peut ainsi passer ~264 k€ par enfant et par période de 15 ans."
	perChild := succParents * (succAbattementDirect + succDonFamilial)
	if perChild != 263730 {
		t.Errorf("per child and per 15-year period: %.0f EUR, the article says ~264 k€", perChild)
	}
	// "Un couple avec deux enfants transmet ainsi 400 000 € tous les 15 ans."
	if got := succParents * succEnfants * succAbattementDirect; got != 400000 {
		t.Errorf("allowances alone: %.0f EUR, the article says 400 000 €", got)
	}
}

func TestRechargeTracksReachThePlateTotals(t *testing.T) {
	for _, tc := range []struct {
		start int
		ages  []int
		total float64
	}{
		{55, []int{55, 70, 85}, 1454920}, // 527 460 + 527 460 + 400 000
		{65, []int{65, 80}, 927460},      // 527 460 + 400 000
		{75, []int{75, 90}, 927460},      // 527 460 + 400 000, the last one at 90 sharp
	} {
		ages, cum := succCumul(tc.start, succHorizon)
		if !reflect.DeepEqual(ages, tc.ages) {
			t.Errorf("start at %d: waves %v, the plate draws %v", tc.start, ages, tc.ages)
		}
		if got := cum[len(cum)-1]; got != tc.total {
			t.Errorf("start at %d: %.0f EUR by %d, the plate says %.0f", tc.start, got, succHorizon, tc.total)
		}
	}
}

// "Commencer à 55 ans plutôt qu'à 75 double le nombre de recharges avant 90
// ans, deux contre une, et l'unique recharge du départ tardif tombe à 90 ans
// tout juste." Everything in that sentence, checked.
func TestRechargeDoublingClaim(t *testing.T) {
	early, _ := succCumul(55, succHorizon)
	late, _ := succCumul(75, succHorizon)
	// A recharge is a wave after the first one.
	earlyReloads, lateReloads := len(early)-1, len(late)-1
	if earlyReloads != 2 || lateReloads != 1 {
		t.Fatalf("reloads by %d: %d starting at 55, %d starting at 75; the article says two against one",
			succHorizon, earlyReloads, lateReloads)
	}
	if earlyReloads != 2*lateReloads {
		t.Errorf("%d reloads against %d: the article claims a doubling", earlyReloads, lateReloads)
	}
	// The claim is knife-edge and the article says so: the late household's
	// only recharge lands exactly on the last year of the frame, and the plate
	// draws it dashed for that reason.
	if late[len(late)-1] != succHorizon {
		t.Errorf("the late reload is at %d, the article and the plate put it at %d tout juste",
			late[len(late)-1], succHorizon)
	}
	if early[len(early)-1] >= succHorizon {
		t.Errorf("the early household's last wave is at %d, it should fall before %d",
			early[len(early)-1], succHorizon)
	}
}

func TestRechargePlateLabels(t *testing.T) {
	svg := figEscalierRecharges()
	for _, want := range []string{
		"1 455 k€", "927 k€", "527 k€", // the three totals
		"+527", "+400", // the treads, equal but for the ones after 80
		"3 vagues", "2 vagues", "1 vague",
		"art. 779 CGI", "art. 790 G", "art. 784 CGI",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	if strings.Contains(svg, "rgba(") {
		t.Error("rgba fills are painted solid black by crengine")
	}
}
