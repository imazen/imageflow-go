package imageflow

import (
	"encoding/json"
	"os"
	"testing"
)

// ---------------------------------------------------------------------------
// Low-level job tests
// ---------------------------------------------------------------------------

func TestJob(t *testing.T) {
	job, err := newJob()
	if err != nil {
		t.Fatal(err)
	}
	defer job.CleanUp()

	data, err := os.ReadFile("image.jpg")
	if err != nil {
		t.Fatal(err)
	}
	command := []byte(`{"io":[{"io_id":0,"direction":"in","io":"placeholder"},{"io_id":1,"direction":"out","io":"placeholder"}],"framewise":{"steps":[{"decode":{"io_id":0}},{"constrain":{"mode":"within","w":400}},"rotate_90",{"encode":{"io_id":1,"preset":{"pngquant":{"quality":80}}}}]}}`)

	if err := job.AddInput(0, data); err != nil {
		t.Fatal(err)
	}
	if err := job.AddOutput(1); err != nil {
		t.Fatal(err)
	}
	if err := job.Message(command); err != nil {
		t.Fatal(err)
	}
	result, err := job.GetOutput(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestJobDoubleCleanUp(t *testing.T) {
	job, err := newJob()
	if err != nil {
		t.Fatal(err)
	}
	job.CleanUp()
	job.CleanUp() // should not panic or double-free
}

func TestJobCheckErrorNilContext(t *testing.T) {
	j := &job{inner: nil}
	if !j.CheckError() {
		t.Error("expected CheckError to return true for nil context")
	}
}

func TestError(t *testing.T) {
	job, err := newJob()
	if err != nil {
		t.Fatal(err)
	}
	defer job.CleanUp()

	data, err := os.ReadFile("image.jpg")
	if err != nil {
		t.Fatal(err)
	}
	// Intentionally malformed JSON (missing opening brace)
	command := []byte(`"io":[{"io_id":0,"direction":"in","io":"placeholder"}],"framewise":{"steps":[{"decode":{"io_id":0}}]}`)

	if err := job.AddInput(0, data); err != nil {
		t.Fatal(err)
	}
	if err := job.AddOutput(1); err != nil {
		t.Fatal(err)
	}
	err = job.Message(command)
	if err == nil {
		t.Error("expected error from malformed JSON command")
	}
}

// ---------------------------------------------------------------------------
// Encoding format tests
// ---------------------------------------------------------------------------

func loadTestImage(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("image.jpg")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestEncodeMozJPEG(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(200).
		Encode(GetBuffer("out"), MozJPEG{Quality: 75, Progressive: true}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	out := m["out"]
	if len(out) == 0 {
		t.Fatal("expected non-empty JPEG output")
	}
	// JPEG magic bytes: FF D8 FF
	if out[0] != 0xFF || out[1] != 0xD8 {
		t.Error("output does not start with JPEG magic bytes")
	}
}

func TestEncodeMozJPEGDefaultQuality(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("expected non-empty output with default quality")
	}
}

func TestEncodeLosslessPNG(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Encode(GetBuffer("out"), LosslessPNG{MaxDeflate: true}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	out := m["out"]
	if len(out) < 8 {
		t.Fatal("expected non-empty PNG output")
	}
	// PNG magic: 89 50 4E 47
	if out[0] != 0x89 || out[1] != 0x50 || out[2] != 0x4E || out[3] != 0x47 {
		t.Error("output does not start with PNG magic bytes")
	}
}

func TestEncodeLossyPNG(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Encode(GetBuffer("out"), LossyPNG{Quality: 80, MinimumQuality: 20, Speed: 5}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("expected non-empty lossy PNG output")
	}
}

func TestEncodeWebP(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Encode(GetBuffer("out"), WebP{Quality: 80}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	out := m["out"]
	if len(out) < 12 {
		t.Fatal("expected non-empty WebP output")
	}
	// WebP magic: RIFF....WEBP
	if string(out[0:4]) != "RIFF" || string(out[8:12]) != "WEBP" {
		t.Error("output does not contain WebP magic bytes")
	}
}

func TestEncodeWebPLossless(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Encode(GetBuffer("out"), WebPLossless{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	out := m["out"]
	if len(out) < 12 {
		t.Fatal("expected non-empty WebP lossless output")
	}
	if string(out[0:4]) != "RIFF" || string(out[8:12]) != "WEBP" {
		t.Error("output does not contain WebP magic bytes")
	}
}

func TestEncodeGIF(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Encode(GetBuffer("out"), GIF{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	out := m["out"]
	if len(out) < 6 {
		t.Fatal("expected non-empty GIF output")
	}
	// GIF magic: GIF89a or GIF87a
	if string(out[0:3]) != "GIF" {
		t.Error("output does not start with GIF magic bytes")
	}
}

// Shorthand encode helpers
func TestPNGShorthand(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		PNG(GetBuffer("out")).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("PNG shorthand produced empty output")
	}
}

func TestJPEGShorthand(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		JPEG(GetBuffer("out")).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("JPEG shorthand produced empty output")
	}
}

func TestWebPShorthand(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		WebP(GetBuffer("out")).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("WebP shorthand produced empty output")
	}
}

func TestGIFShorthand(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		GIF(GetBuffer("out")).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("GIF shorthand produced empty output")
	}
}

// ---------------------------------------------------------------------------
// Constraint tests
// ---------------------------------------------------------------------------

func TestConstrainWithin(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithin(150, 150).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("ConstrainWithin produced empty output")
	}
}

func TestConstrainWithinW(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(200).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("ConstrainWithinW produced empty output")
	}
}

func TestConstrainWithinH(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinH(200).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("ConstrainWithinH produced empty output")
	}
}

func TestConstrainFull(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Constrain(Constrain{
			Mode: "within",
			W:    300,
			H:    300,
			Hint: ConstraintHint{
				SharpenPercent:    float32(10),
				DownFilter:        "lanczos",
				UpFilter:          "ginseng",
				ScalingColorspace: "linear",
			},
		}).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Constrain with hints produced empty output")
	}
}

func TestConstrainWithGravityAndCanvasColor(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Constrain(Constrain{
			Mode:        "within",
			W:           300,
			H:           300,
			Gravity:     ConstraintGravity{X: 50, Y: 50},
			CanvasColor: Black{},
		}).
		Encode(GetBuffer("out"), LosslessPNG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Constrain with gravity produced empty output")
	}
}

// ---------------------------------------------------------------------------
// Transform tests
// ---------------------------------------------------------------------------

func TestRotate90(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Rotate90().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Rotate90 produced empty output")
	}
}

func TestRotate180(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Rotate180().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Rotate180 produced empty output")
	}
}

func TestRotate270(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Rotate270().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Rotate270 produced empty output")
	}
}

func TestFlipH(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		FlipH().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("FlipH produced empty output")
	}
}

func TestFlipV(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		FlipV().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("FlipV produced empty output")
	}
}

func TestAllRotationsAndFlips(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Rotate90().FlipH().Rotate180().FlipV().Rotate270().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("chained transforms produced empty output")
	}
}

// ---------------------------------------------------------------------------
// Region / crop tests
// ---------------------------------------------------------------------------

func TestRegion(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Region(Region{
			X1: 10, Y1: 10, X2: 200, Y2: 200,
			BackgroundColor: Black{},
		}).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Region produced empty output")
	}
}

func TestRegionPercentage(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		RegionPercentage(RegionPercentage{
			X1: 10, Y1: 10, X2: 90, Y2: 90,
			BackgroundColor: Black{},
		}).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("RegionPercentage produced empty output")
	}
}

func TestCropWhitespace(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		CropWhitespace(80, 0.5).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("CropWhitespace produced empty output")
	}
}

// ---------------------------------------------------------------------------
// Drawing / canvas tests
// ---------------------------------------------------------------------------

func TestFillRect(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		FillRect(0, 0, 50, 50, Black{}).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("FillRect produced empty output")
	}
}

func TestExpandCanvas(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(200).
		ExpandCanvas(ExpandCanvas{
			Left: 10, Right: 10, Top: 20, Bottom: 20,
			Color: Black{},
		}).
		Encode(GetBuffer("out"), LosslessPNG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("ExpandCanvas produced empty output")
	}
}

func TestExpandCanvasTransparent(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(200).
		ExpandCanvas(ExpandCanvas{
			Left: 5, Right: 5, Top: 5, Bottom: 5,
			Color: Transparent(""),
		}).
		Encode(GetBuffer("out"), LosslessPNG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("ExpandCanvas with Transparent produced empty output")
	}
}

// ---------------------------------------------------------------------------
// Color filter tests
// ---------------------------------------------------------------------------

func TestGrayscaleNTSC(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		GrayscaleNTSC().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("GrayscaleNTSC produced empty output")
	}
}

func TestGrayscaleFlat(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		GrayscaleFlat().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("GrayscaleFlat produced empty output")
	}
}

func TestGrayscaleBT709(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		GrayscaleBT709().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("GrayscaleBT709 produced empty output")
	}
}

func TestGrayscaleRY(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		GrayscaleRY().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("GrayscaleRY produced empty output")
	}
}

func TestSepia(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Sepia().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Sepia produced empty output")
	}
}

func TestInvert(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Invert().
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Invert produced empty output")
	}
}

func TestAlpha(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Alpha(0.5).
		Encode(GetBuffer("out"), LosslessPNG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Alpha produced empty output")
	}
}

func TestContrast(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Contrast(1.2).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Contrast produced empty output")
	}
}

func TestBrightness(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Brightness(0.8).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Brightness produced empty output")
	}
}

func TestSaturation(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Saturation(0.5).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Saturation produced empty output")
	}
}

func TestWhiteBalanceSRGB(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		WhiteBalanceSRGB(80).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("WhiteBalanceSRGB produced empty output")
	}
}

// ---------------------------------------------------------------------------
// Branching and multi-output tests
// ---------------------------------------------------------------------------

func TestBranch(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Branch(func(s *Steps) {
			s.ConstrainWithin(100, 100).
				Encode(GetBuffer("small"), MozJPEG{})
		}).
		ConstrainWithin(400, 400).
		Encode(GetBuffer("large"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["small"]) == 0 {
		t.Error("branch 'small' produced empty output")
	}
	if len(m["large"]) == 0 {
		t.Error("branch 'large' produced empty output")
	}
	if len(m["small"]) >= len(m["large"]) {
		t.Error("expected small output to be smaller than large output")
	}
}

func TestMultipleBranches(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Branch(func(s *Steps) {
			s.ConstrainWithinW(100).
				Encode(GetBuffer("thumb"), MozJPEG{Quality: 60})
		}).
		Branch(func(s *Steps) {
			s.ConstrainWithinW(300).
				GrayscaleFlat().
				Encode(GetBuffer("gray"), MozJPEG{})
		}).
		ConstrainWithinW(500).
		Encode(GetBuffer("full"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"thumb", "gray", "full"} {
		if len(m[key]) == 0 {
			t.Errorf("branch %q produced empty output", key)
		}
	}
}

func TestBranchNestedRegion(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(400).
		Branch(func(s *Steps) {
			s.ConstrainWithin(100, 100).
				Rotate180().
				Region(Region{
					X1: 0, Y1: 0, X2: 80, Y2: 80,
					BackgroundColor: Black{},
				}).
				Branch(func(s2 *Steps) {
					s2.GrayscaleFlat().Encode(GetBuffer("gray_crop"), MozJPEG{})
				}).
				Encode(GetBuffer("crop"), MozJPEG{})
		}).
		Encode(GetBuffer("main"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"gray_crop", "crop", "main"} {
		if len(m[key]) == 0 {
			t.Errorf("branch %q produced empty output", key)
		}
	}
}

// ---------------------------------------------------------------------------
// I/O operation tests
// ---------------------------------------------------------------------------

func TestIOOperation(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).ConstrainWithinW(400).Branch(func(step *Steps) {
		step.ConstrainWithin(100, 100).Region(Region{
			X1: 0, Y1: 0, X2: 200, Y2: 200,
			BackgroundColor: Black{},
		}).Rotate180().Encode(GetBuffer("key_1"), MozJPEG{})
	}).Encode(GetBuffer("key_2"), MozJPEG{}).Execute()
	if err != nil {
		t.Fatal(err)
	}
	if m["key_1"] == nil || m["key_2"] == nil {
		t.Error("Buffer failed")
	}
}

func TestFileIO(t *testing.T) {
	data := loadTestImage(t)
	outPath := "test_file_io_output.jpg"
	defer os.Remove(outPath)

	step := NewStep()
	_, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(200).
		Encode(NewFile(outPath), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestFileInputAndOutput(t *testing.T) {
	outPath := "test_file_roundtrip.png"
	defer os.Remove(outPath)

	step := NewStep()
	_, err := step.Decode(NewFile("image.jpg")).
		ConstrainWithinW(100).
		Encode(NewFile(outPath), LosslessPNG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

// ---------------------------------------------------------------------------
// Watermark tests
// ---------------------------------------------------------------------------

func TestWatermark(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Watermark(NewBuffer(data), ConstraintGravity{X: 100, Y: 100}, "", nil, .5, nil).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Watermark produced empty output")
	}
}

func TestWatermarkWithPercentageFitBox(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Watermark(NewBuffer(data), nil, "within", PercentageFitBox{
			X1: 0, Y1: 0, X2: 50, Y2: 50,
		}, 0.3, nil).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Watermark with PercentageFitBox produced empty output")
	}
}

func TestWatermarkWithMarginFitBox(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(400).
		Watermark(NewBuffer(data), nil, "within", MarginFitBox{
			Left: 10, Right: 10, Top: 10, Bottom: 10,
		}, 0.5, nil).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Watermark with MarginFitBox produced empty output")
	}
}

// ---------------------------------------------------------------------------
// DrawExact / CopyRectangle tests
// ---------------------------------------------------------------------------

func TestDrawExact(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(400).
		DrawExact(func(s *Steps) {
			s.Decode(NewBuffer(data))
		}, DrawExact{
			X: 0, Y: 0, W: 100, H: 100,
			Blend: "overwrite",
		}).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("DrawExact produced empty output")
	}
}

func TestCopyRectangle(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(400).
		CopyRectangle(func(s *Steps) {
			s.Decode(NewBuffer(data))
		}, RectangleToCanvas{
			FromX: 0, FromY: 0, W: 50, H: 50,
			X: 10, Y: 10,
		}).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("CopyRectangle produced empty output")
	}
}

// ---------------------------------------------------------------------------
// Command string test
// ---------------------------------------------------------------------------

func TestCommand(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		Command("width=200&height=200&mode=max").
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("Command produced empty output")
	}
}

// ---------------------------------------------------------------------------
// JSON serialization test
// ---------------------------------------------------------------------------

func TestToJSON(t *testing.T) {
	step := NewStep()
	step.Decode(NewBuffer([]byte{})).
		ConstrainWithinW(200).
		Rotate90().
		Encode(GetBuffer("out"), MozJPEG{})
	js := step.ToJSON()
	if len(js) == 0 {
		t.Fatal("ToJSON produced empty output")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(js, &parsed); err != nil {
		t.Fatalf("ToJSON produced invalid JSON: %v", err)
	}
	framewise, ok := parsed["framewise"]
	if !ok {
		t.Fatal("JSON missing 'framewise' key")
	}
	fw := framewise.(map[string]interface{})
	graph, ok := fw["graph"]
	if !ok {
		t.Fatal("JSON missing 'graph' key")
	}
	g := graph.(map[string]interface{})
	if _, ok := g["nodes"]; !ok {
		t.Fatal("JSON missing 'nodes' key")
	}
	if _, ok := g["edges"]; !ok {
		t.Fatal("JSON missing 'edges' key")
	}
}

// ---------------------------------------------------------------------------
// Complex pipeline (original TestStep equivalent)
// ---------------------------------------------------------------------------

func TestStep(t *testing.T) {
	step := NewStep()
	_, errorInStep := step.Decode(NewURL("https://jpeg.org/images/jpeg2000-home.jpg")).FlipV().Watermark(NewFile("image.jpg"), nil, "within", PercentageFitBox{
		X1: 0,
		Y1: 0,
		X2: 50,
		Y2: 50,
	}, 0.3, nil).ConstrainWithinW(400).FillRect(0, 0, 8, 8, Black{}).Branch(func(step *Steps) {
		step.ConstrainWithin(100, 100).Rotate180().Region(Region{
			X1:              0,
			Y1:              0,
			X2:              200,
			Y2:              200,
			BackgroundColor: Black{},
		}).Branch(func(step *Steps) {
			step.GrayscaleFlat().Encode(NewFile("gray_small.jpg"), MozJPEG{})
		}).Encode(NewFile("small.jpg"), MozJPEG{})
	}).DrawExact(func(steps *Steps) {
		steps.Decode(NewFile("image.jpg"))
	}, DrawExact{
		X:     0,
		Y:     0,
		W:     100,
		H:     100,
		Blend: "overwrite",
	}).ExpandCanvas(ExpandCanvas{Top: 10, Color: Black{}}).Encode(NewFile("medium.jpg"), MozJPEG{}).Execute()
	if errorInStep != nil {
		t.Error(errorInStep)
		t.FailNow()
	}
}

// ---------------------------------------------------------------------------
// Multi-format output from single decode
// ---------------------------------------------------------------------------

func TestMultiFormatOutput(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Branch(func(s *Steps) {
			s.Encode(GetBuffer("jpeg"), MozJPEG{Quality: 70})
		}).
		Branch(func(s *Steps) {
			s.Encode(GetBuffer("png"), LosslessPNG{})
		}).
		Branch(func(s *Steps) {
			s.Encode(GetBuffer("webp"), WebP{Quality: 80})
		}).
		Encode(GetBuffer("gif"), GIF{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"jpeg", "png", "webp", "gif"} {
		if len(m[key]) == 0 {
			t.Errorf("format %q produced empty output", key)
		}
	}
}

// ---------------------------------------------------------------------------
// Chained color filters
// ---------------------------------------------------------------------------

func TestChainedColorFilters(t *testing.T) {
	data := loadTestImage(t)
	step := NewStep()
	m, err := step.Decode(NewBuffer(data)).
		ConstrainWithinW(100).
		Brightness(0.9).
		Contrast(1.1).
		Saturation(0.8).
		Encode(GetBuffer("out"), MozJPEG{}).
		Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(m["out"]) == 0 {
		t.Fatal("chained color filters produced empty output")
	}
}

// ---------------------------------------------------------------------------
// Benchmark
// ---------------------------------------------------------------------------

func BenchmarkSteps(b *testing.B) {
	data, err := os.ReadFile("image.jpg")
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		step := NewStep()
		step.Decode(NewBuffer(data)).ConstrainWithinW(400).Branch(func(step *Steps) {
			step.ConstrainWithin(100, 100).Region(Region{
				X1: 0, Y1: 0, X2: 200, Y2: 200,
				BackgroundColor: Black{},
			}).Rotate180().Encode(&Buffer{}, MozJPEG{})
		}).Encode(&Buffer{}, MozJPEG{}).Execute()
	}
}
