// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workflow

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/dg-pz-gocommon/elasticsearch"
	"github.com/venicegeo/dg-pz-gocommon/gocommon"
)

type ClientTester struct {
	suite.Suite
	client *Client
	sys    *piazza.SystemConfig
}

func (suite *ClientTester) SetupSuite() {
	assertNoData(suite.T(), suite.client)
}

func (suite *ClientTester) TearDownSuite() {
	assertNoData(suite.T(), suite.client)
}

//---------------------------------------------------------------------------

func (suite *ClientTester) Test11Admin() {
	t := suite.T()
	assert := assert.New(t)

	client := suite.client

	_, err := client.GetStats()
	assert.NoError(err)
}

func (suite *ClientTester) Test12AlertResource() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error

	a1 := Alert{TriggerID: "dummyT1", EventID: "dummyE1"}
	respAlert, err := client.PostAlert(&a1)
	id := respAlert.AlertID
	assert.NoError(err)

	alerts, err := client.GetAllAlerts(100, 0)
	assert.NoError(err)
	assert.Len(*alerts, 1)
	assert.EqualValues(id, (*alerts)[0].AlertID)
	assert.EqualValues("dummyT1", (*alerts)[0].TriggerID)
	assert.EqualValues("dummyE1", (*alerts)[0].EventID)

	alert, err := client.GetAlert(id)
	assert.NoError(err)
	assert.EqualValues(id, alert.AlertID)

	alert, err = client.GetAlert("nosuchalert1")
	assert.Error(err)

	err = client.DeleteAlert("nosuchalert2")
	assert.Error(err)

	err = client.DeleteAlert(id)
	assert.NoError(err)

	alert, err = client.GetAlert(id)
	assert.Error(err)

	alerts, err = client.GetAllAlerts(100, 0)
	assert.NoError(err)
	assert.Len(*alerts, 0)
}

func (suite *ClientTester) Test13EventResource() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error

	mapping := map[string]interface{}{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventTypeName := "mytype"
	eventType := &EventType{Name: eventTypeName, Mapping: mapping}
	respEventType, err := client.PostEventType(eventType)
	etID := respEventType.EventTypeID
	assert.NoError(err)
	defer func() {
		err = client.DeleteEventType(etID)
		assert.NoError(err)
	}()

	event := &Event{
		EventTypeID: etID,
		CreatedOn:   piazza.NewTimeStamp(),
		Data: map[string]interface{}{
			"myint": 17,
			"mystr": "quick",
		},
	}
	respEvent, err := client.PostEvent(event)
	eID := respEvent.EventID
	assert.NoError(err)

	defer func() {
		err = client.DeleteEvent(eID)
		assert.NoError(err)
	}()

	//events, err := workflow.GetAllEvents("")
	//assert.NoError(err)
	//assert.Len(*events, 1)
	//assert.EqualValues(eID, (*events)[0].ID)

	tmp, err := client.GetEvent(eID)
	assert.NoError(err)
	assert.EqualValues(eID, tmp.EventID)
}

func (suite *ClientTester) Test14EventTypeResource() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error

	mapping := map[string]interface{}{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventType := &EventType{Name: "typnam", Mapping: mapping}
	respEventType, err := client.PostEventType(eventType)
	id := respEventType.EventTypeID
	defer func() {
		err = client.DeleteEventType(id)
		assert.NoError(err)
	}()
	_, _ = client.GetEventTypeByName("typnam")

	eventTypes, err := client.GetAllEventTypes(100, 0)
	assert.NoError(err)
	assert.Len(*eventTypes, 3)

	// Find the right event type and assert
	var theType EventType
	for _, eventtyp := range *eventTypes {
		if eventtyp.EventTypeID == id {
			theType = eventtyp
		}
	}
	assert.EqualValues(id, theType.EventTypeID)

	tmp, err := client.GetEventType(id)
	assert.NoError(err)
	assert.EqualValues(id, tmp.EventTypeID)
}

func (suite *ClientTester) Test15One() {

	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var eventTypeName = "EventTypeA"

	var etID piazza.Ident
	{
		mapping := map[string]interface{}{
			"num": elasticsearch.MappingElementTypeInteger,
			"str": elasticsearch.MappingElementTypeString,
		}

		eventType := &EventType{Name: eventTypeName, Mapping: mapping}

		respEventType, err := client.PostEventType(eventType)
		etID = respEventType.EventTypeID
		assert.NoError(err)

		defer func() {
			err := client.DeleteEventType(etID)
			assert.NoError(err)
		}()
	}

	var tID piazza.Ident
	{
		x1 := &Trigger{
			Name:        "the x1 trigger",
			EventTypeID: etID,
			Condition: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"num": 17,
					},
				},
			},
			Job: JobRequest{
				CreatedBy: "test",
				JobType: JobType{
					Type: "execute-service",
					Data: map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}

		respTrigger, err := client.PostTrigger(x1)
		tID = respTrigger.TriggerID
		assert.NoError(err)

		defer func() {
			err := client.DeleteTrigger(tID)
			assert.NoError(err)
		}()
	}

	var e1ID piazza.Ident
	{
		// will cause trigger t1ID
		e1 := &Event{
			EventTypeID: etID,
			CreatedOn:   piazza.NewTimeStamp(),
			Data: map[string]interface{}{
				"num":      17,
				"str":      "quick",
				"userName": "my-api-key-38n987",
			},
		}

		respEvent1, err := client.PostEvent(e1)
		e1ID = respEvent1.EventID
		assert.NoError(err)

		defer func() {
			err := client.DeleteEvent(e1ID)
			assert.NoError(err)
		}()
	}

	var e2ID piazza.Ident
	{
		// will cause no triggers
		e2 := &Event{
			EventTypeID: etID,
			CreatedOn:   piazza.NewTimeStamp(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
			},
		}

		respEvent2, err := client.PostEvent(e2)
		e2ID = respEvent2.EventID
		assert.NoError(err)

		defer func() {
			err := client.DeleteEvent(e2ID)
			assert.NoError(err)
		}()
	}
}

func (suite *ClientTester) Test16TriggerResource() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error

	mapping := map[string]interface{}{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventType := &EventType{Name: "typnam2", Mapping: mapping}
	respEventType, err := client.PostEventType(eventType)
	etID := respEventType.EventTypeID

	defer func() {
		err = client.DeleteEventType(etID)
		assert.NoError(err)
	}()

	t1 := Trigger{
		Name:        "the x1 trigger",
		EventTypeID: etID,
		Condition: map[string]interface{}{
			"match": map[string]interface{}{
				"myint": 17,
			},
		},
		Job: JobRequest{
			CreatedBy: "test",
			JobType: JobType{
				Type: "execute-service",
				Data: map[string]interface{}{
					// "dataInputs": map[string]interface{},
					// "dataOutput": map[string]interface{},
					"serviceId": "ddd5134",
				},
			},
		},
	}
	respTrigger, err := client.PostTrigger(&t1)
	t1ID := respTrigger.TriggerID
	assert.NoError(err)

	defer func() {
		err = client.DeleteTrigger(t1ID)
		assert.NoError(err)
	}()

	tmp, err := client.GetTrigger(t1ID)
	assert.NoError(err)
	assert.EqualValues(t1ID, tmp.TriggerID)

	triggers, err := client.GetAllTriggers(100, 0)
	assert.NoError(err)
	assert.Len(*triggers, 1)
	assert.EqualValues(t1ID, (*triggers)[0].TriggerID)
}

func (suite *ClientTester) Test17Triggering() {

	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	//-----------------------------------------------------

	var etC, etD, etE piazza.Ident
	{
		mapping := map[string]interface{}{
			"num": elasticsearch.MappingElementTypeInteger,
			"str": elasticsearch.MappingElementTypeString,
		}
		eventTypeC := &EventType{Name: "EventType C", Mapping: mapping}
		eventTypeD := &EventType{Name: "EventType D", Mapping: mapping}
		eventTypeE := &EventType{Name: "EventType E", Mapping: mapping}
		respEventTypeC, err := client.PostEventType(eventTypeC)
		etC = respEventTypeC.EventTypeID
		assert.NoError(err)
		respEventTypeD, err := client.PostEventType(eventTypeD)
		etD = respEventTypeD.EventTypeID
		assert.NoError(err)
		respEventTypeE, err := client.PostEventType(eventTypeE)
		etE = respEventTypeE.EventTypeID
		assert.NoError(err)

		eventTypes, err := client.GetAllEventTypes(100, 0)
		assert.NoError(err)
		assert.Len(*eventTypes, 5)
	}

	defer func() {
		err := client.DeleteEventType(etC)
		assert.NoError(err)
		err = client.DeleteEventType(etD)
		assert.NoError(err)
		err = client.DeleteEventType(etE)
		assert.NoError(err)
	}()

	////////////////

	var tB piazza.Ident
	{
		t1 := &Trigger{
			Name:        "Trigger A",
			EventTypeID: etC,
			Condition: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"str": "quick",
					},
				},
			},
			Job: JobRequest{
				CreatedBy: "test",
				JobType: JobType{
					Type: "execute-service",
					Data: map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}
		respTriggerA, err := client.PostTrigger(t1)
		tA := respTriggerA.TriggerID
		assert.NoError(err)
		defer func() {
			err = client.DeleteTrigger(tA)
			assert.NoError(err)
		}()

		t2 := &Trigger{
			Name:        "Trigger B",
			EventTypeID: etD,
			Condition: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"num": 18,
					},
				},
			},
			Job: JobRequest{
				CreatedBy: "test",
				JobType: JobType{
					Type: "execute-service",
					Data: map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}
		respTriggerB, err := client.PostTrigger(t2)
		tB = respTriggerB.TriggerID
		assert.NoError(err)
		defer func() {
			err = client.DeleteTrigger(tB)
			assert.NoError(err)
		}()

		triggers, err := client.GetAllTriggers(100, 0)
		assert.NoError(err)
		assert.Len(*triggers, 2)
	}

	var eF, eG, eH piazza.Ident
	{
		// will cause trigger TA
		e1 := Event{
			EventTypeID: etC,
			CreatedOn:   piazza.NewTimeStamp(),
			Data: map[string]interface{}{
				"num":      17,
				"str":      "quick",
				"userName": "my-api-key-38n987",
				"jobId":    "43688858-b6d4-4ef9-a98b-163e1980bba8",
			},
		}
		respEventF, err := client.PostEvent(&e1)
		eF = respEventF.EventID
		assert.NoError(err)
		defer func() {
			err = client.DeleteEvent(eF)
			assert.NoError(err)
		}()

		// will cause trigger TB
		e2 := Event{
			EventTypeID: etD,
			CreatedOn:   piazza.NewTimeStamp(),
			Data: map[string]interface{}{
				"num":      18,
				"str":      "brown",
				"userName": "my-api-key-38n987",
				"jobId":    "43688858-b6d4-4ef9-a98b-163e1980bba8",
			},
		}
		respEventG, err := client.PostEvent(&e2)
		eG = respEventG.EventID
		assert.NoError(err)
		defer func() {
			err = client.DeleteEvent(eG)
			assert.NoError(err)
		}()

		// will cause no triggers
		e3 := Event{
			EventTypeID: etE,
			CreatedOn:   piazza.NewTimeStamp(),
			Data: map[string]interface{}{
				"num": 19,
				"str": "fox",
			},
		}
		respEventH, err := client.PostEvent(&e3)
		eH = respEventH.EventID
		assert.NoError(err)
		defer func() {
			err = client.DeleteEvent(eH)
			assert.NoError(err)
		}()
	}
}
