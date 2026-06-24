package config

import (
	"testing"
)

func TestLoadBillingCatalog(t *testing.T) {
	c, err := LoadBillingCatalog()
	if err != nil {
		t.Fatalf("LoadBillingCatalog failed: %v", err)
	}
	if c.Version == "" {
		t.Error("expected version")
	}
	if len(c.TreatmentFees) == 0 {
		t.Error("expected treatment fees")
	}
	if len(c.NursingFees) == 0 {
		t.Error("expected nursing fees")
	}
	if c.InjectionFee.Price == 0 {
		t.Error("expected injection fee price")
	}
	if len(c.Materials) == 0 {
		t.Error("expected materials")
	}

	tf, ok := c.TreatmentFeeFor("HD")
	if !ok || tf.Price != 399 {
		t.Errorf("HD treatment fee: ok=%v price=%v", ok, tf.Price)
	}

	if _, ok := c.TreatmentFeeFor("UNKNOWN_MODE"); ok {
		t.Error("unknown mode should not be found")
	}

	mp, ok := c.MaterialFor(15)
	if !ok || mp.Name != "RA130" {
		t.Errorf("material 15: ok=%v name=%q", ok, mp.Name)
	}

	if _, ok := c.MaterialFor(99999); ok {
		t.Error("unknown charge item should not be found")
	}

	nf := c.NursingFeeFor("AVF")
	if len(nf) == 0 {
		t.Error("expected AVF nursing fees")
	}

	if nf2 := c.NursingFeeFor("UNKNOWN"); len(nf2) != 0 {
		t.Error("unknown access type should have no fees")
	}

	c2, err := LoadBillingCatalog()
	if err != nil || c2 != c {
		t.Error("LoadBillingCatalog should return singleton")
	}
}
