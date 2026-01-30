package lichess

import (
	"encoding/json"
	"testing"
)

func TestPlayerProfileType(t *testing.T) {

	input := `{"id":"igramnet","username":"igramnet","perfs":{"bullet":{"games":0,"rating":1500,"rd":500,"prog":0,"prov":true},"blitz":{"games":1,"rating":1291,"rd":290,"prog":0,"prov":true},"rapid":{"games":11,"rating":1875,"rd":113,"prog":0,"prov":true},"classical":{"games":0,"rating":1500,"rd":500,"prog":0,"prov":true},"correspondence":{"games":0,"rating":1500,"rd":500,"prog":0,"prov":true}},"createdAt":1736777074253,"seenAt":1765704857689,"playTime":{"total":22624,"tv":0},"url":"https://lichess.org/@/igramnet","count":{"all":26,"rated":12,"draw":2,"loss":6,"win":18,"bookmark":0,"playing":0,"import":0,"me":1},"kid":true,"followable":true,"following":false,"blocking":false}`

	var profile PlayerProfile
	err := json.Unmarshal([]byte(input), &profile)
	if err != nil {
		t.Errorf("Failed to unmarshal PlayerProfile: %v", err)
	}

	// check that we correctly parsed the provisional status of rapid rating
	if !profile.IsProvisional(Rapid) {
		t.Errorf("Expected IsProvisional to be true, got false")
	}

	inputNonProv := `{"id":"stodorov","username":"stodorov","perfs":{"bullet":{"games":0,"rating":1500,"rd":500,"prog":0,"prov":true},"blitz":{"games":21,"rating":1706,"rd":197,"prog":-97,"prov":true},"rapid":{"games":4264,"rating":1888,"rd":45,"prog":6},"classical":{"games":880,"rating":1870,"rd":197,"prog":-6,"prov":true},"correspondence":{"games":0,"rating":1500,"rd":500,"prog":0,"prov":true},"puzzle":{"games":1826,"rating":2017,"rd":77,"prog":0}},"createdAt":1585075810044,"seenAt":1765706199479,"playTime":{"total":5780331,"tv":3503},"url":"https://lichess.org/@/stodorov","playing":"https://lichess.org/pElXEnoe/white","count":{"all":5530,"rated":5166,"draw":439,"loss":2912,"win":2179,"bookmark":0,"playing":1,"import":0,"me":1},"followable":true,"following":false,"blocking":false}`

	var profileNonProv PlayerProfile
	err = json.Unmarshal([]byte(inputNonProv), &profileNonProv)
	if err != nil {
		t.Errorf("Failed to unmarshal PlayerProfile: %v", err)
	}

	// check that we correctly parsed the non-provisional status of rapid rating
	if profileNonProv.IsProvisional(Rapid) {
		t.Errorf("Expected IsProvisional to be false, got true")
	}

}
