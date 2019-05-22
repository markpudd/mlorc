// Assume only one batch queue running on cluster
package mlorc

import (
	"sync"
	"io"
	"encoding/json"
)


var PointsMux sync.Mutex

type DataPoint  struct {
  Epoch int
  Val_loss float64
  Val_acc float64
  Loss float64
  Acc float64
}

type DataLogger struct {
	MaxEpoch int
  Points map[string] []DataPoint
	Colours map[string] string
  PointsMux sync.Mutex
}

var colourList [6]string

func NewDataLogger() *DataLogger {
	dl := new(DataLogger)
  dl.MaxEpoch = 0
	dl.Points = make(map[string][]DataPoint)
	dl.Colours = make(map[string]string)
	colourList = [6]string{"#28D094",
													"#D0D033",
													"#094094",
													"#123445",
													"#999999",
													"#940011"	}
  return dl;
}



func (dl *DataLogger)AddPoint(data io.ReadCloser, name string) {
  var point DataPoint
  decoder := json.NewDecoder(data)
  err := decoder.Decode(&point)
  if err != nil {
      panic(err)
  }
  dl.PointsMux.Lock()
	points, ok := dl.Points[name]
	if !ok {
			points = make([]DataPoint,0)
			dl.Colours[name] = colourList[len(dl.Points) % 6]
	}
  points = append(points,point)
	dl.Points[name] =points


	if point.Epoch >= dl.MaxEpoch {
		dl.MaxEpoch = point.Epoch+1
	}
  dl.PointsMux.Unlock()
}

func (dl *DataLogger) GetJson( tp string) []byte {


	type Dataset struct {
		Label string `json:"label"`
		Colour string `json:"borderColor"`
		Fill bool `json:"fill"`
		Data []float64 `json:"data"`
	}

  type JData struct {
    Labels []int `json:"labels"`
    Datasets []Dataset  `json:"datasets"`
  }

	datasets := []Dataset{}
	labels := make([]int,dl.MaxEpoch)
	for i := 0; i < dl.MaxEpoch; i++ {
		labels[i] = i
	}


  dl.PointsMux.Lock()
	for key := range dl.Points {
		dataset := Dataset {}
		dataset.Label = key
		dataset.Colour = dl.Colours[key]
		dataset.Fill = false
		data := make([]float64, len(dl.Points[key]))
	  for _,point := range dl.Points[key] {
			if tp == "Loss" {
	    	data[point.Epoch] = point.Loss
			} else if tp == "Acc" {
	    	data[point.Epoch] = point.Acc
			} else if tp == "Val_loss" {
	    	data[point.Epoch] = point.Val_acc
			} else {
				data[point.Epoch] = point.Val_loss
			}
		}
		dataset.Data = data
		datasets = append(datasets,dataset)
  }
  jdata := JData{
    Labels: labels,
    Datasets: datasets,
  }
  b, err := json.Marshal(jdata)
  dl.PointsMux.Unlock()
   if err != nil {
      panic(err)
   }

   return b;
}
