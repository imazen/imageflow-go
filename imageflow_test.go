package imageflow

import (
	"io/ioutil"
	"testing"
)

func TestJob(t *testing.T) {
	job := newJob()
	data, _ := ioutil.ReadFile("image.jpg")
	command := []byte("{\"io\":[{\"io_id\":0,\"direction\":\"in\",\"io\":\"placeholder\"},{\"io_id\":1,\"direction\":\"out\",\"io\":\"placeholder\"}],\"framewise\":{\"steps\":[{\"decode\":{\"io_id\":0}},{\"constrain\":{\"mode\":\"within\",\"w\":400}},\"rotate_90\",{\"encode\":{\"io_id\":1,\"preset\":{\"pngquant\":{\"quality\":80}}}}]}}")

	job.AddInput(0, data)
	job.AddOutput(1)
	err := job.Message(command)
	result, _ := job.GetOutput(1)
	ioutil.WriteFile("./output.jpg", result, 0644)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestStep(t *testing.T) {
	step := NewStep()
	_, errorInStep := step.Decode(&URL{URL: "https://jpeg.org/images/jpeg2000-home.jpg"}).FlipV().Watermark(&File{Filename: "image.jpg"}, nil, "within", PercentageFitBox{
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
			step.GrayscaleFlat().Encode(&File{Filename: "gray_small.jpg"}, MozJPEG{})
		}).Encode(&File{Filename: "small.jpg"}, MozJPEG{})
	}).DrawExact(func(steps *Steps) {
		steps.Decode(&File{Filename: "image.jpg"})
	}, DrawExact{
		X:     0,
		Y:     0,
		W:     100,
		H:     100,
		Blend: "overwrite",
	}).ExpandCanvas(ExpandCanvas{Top: 10, Color: Black{}}).Encode(&File{Filename: "medium.jpg"}, MozJPEG{}).Execute()
	if errorInStep != nil {
		t.Error(errorInStep)
		t.FailNow()
	}

}

func BenchmarkSteps(b *testing.B) {
	data, _ := ioutil.ReadFile("image.jpg")
	for i := 0; i < b.N; i++ {
		step := NewStep()
		step.Decode(&Buffer{Buffer: data}).ConstrainWithinW(400).Branch(func(step *Steps) {
			step.ConstrainWithin(100, 100).Region(Region{
				X1:              0,
				Y1:              0,
				X2:              200,
				Y2:              200,
				BackgroundColor: Black{},
			}).Rotate180().Encode(&Buffer{}, MozJPEG{})
		}).Encode(&Buffer{}, MozJPEG{}).Execute()
	}
}

func TestError(t *testing.T) {
	job := newJob()
	data, _ := ioutil.ReadFile("image.jpg")
	command := []byte("\"io\":[{\"io_id\":0,\"direction\":\"in\",\"io\":\"placeholder\"},{\"io_id\":1,\"direction\":\"out\",\"io\":\"placeholder\"}],\"framewise\":{\"steps\":[{\"decode\":{\"io_id\":0}},{\"constrain\":{\"mode\":\"within\",\"w\":400}},\"rotate_90\",{\"encode\":{\"io_id\":1,\"preset\":{\"pngquant\":{\"quality\":80}}}}]}}")

	job.AddInput(0, data)
	job.AddOutput(1)
	err := job.Message(command)
	result, _ := job.GetOutput(1)
	ioutil.WriteFile("./output.jpg", result, 0644)
	if err == nil {
		t.Error("Error should not be null")
		t.Fail()
	}
}
