/*
Copyright © 2020 Dell Inc. or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "github.com/dell/csi-baremetal/api/v1"
	accrd "github.com/dell/csi-baremetal/api/v1/availablecapacitycrd"
	"github.com/dell/csi-baremetal/api/v1/volumecrd"
	errTypes "github.com/dell/csi-baremetal/pkg/base/error"
)

func setup() *CRHelper {
	k, err := GetFakeKubeClient(testNs, testLogger)
	if err != nil {
		panic(err)
	}
	return NewCRHelper(k, testLogger)
}

func TestCRHelper_GetACByLocation(t *testing.T) {
	ch := setup()
	expectedAC := testACCR
	err := ch.k8sClient.CreateCR(testCtx, expectedAC.Name, &expectedAC)
	assert.Nil(t, err)

	currentAC, err := ch.GetACByLocation(testACCR.Spec.Location)
	assert.Nil(t, err)
	assert.Equal(t, expectedAC.Spec, currentAC.Spec)

	// expected nil because of empty string as a location
	currentAC, err = ch.GetACByLocation("")
	assert.Equal(t, err, errTypes.ErrorNotFound)
}

func TestCRHelper_GetVolumeByLocation(t *testing.T) {
	ch := setup()
	expectedV := testVolumeCR.DeepCopy()
	err := ch.k8sClient.CreateCR(testCtx, expectedV.Name, expectedV)
	assert.Nil(t, err)
	ctx := context.Background()
	currentVols, _ := ch.GetVolumesByLocation(ctx, expectedV.Spec.Location)
	assert.NotEmpty(t, currentVols)
	assert.Equal(t, expectedV.Spec, currentVols[0].Spec)

	// expected nil because of empty string as a location
	currentVols, _ = ch.GetVolumesByLocation(ctx, "")
	assert.Nil(t, currentVols)

	// lvm
	ch = setup()
	testVolume := testVolumeCR.DeepCopy()
	testVolume.Spec.Location = testLVGCR.Name
	testVolume.Spec.LocationType = v1.LocationTypeLVM
	err = ch.k8sClient.CreateCR(testCtx, testVolume.Name, testVolume)
	assert.Nil(t, err)
	testLVGCR1 := testLVGCR.DeepCopy()
	err = ch.k8sClient.CreateCR(testCtx, testLVGCR.Name, testLVGCR1)
	assert.Nil(t, err)
	currentVols, _ = ch.GetVolumesByLocation(ctx, testDriveLocation1)
	assert.NotEmpty(t, currentVols)
}

func TestCRHelper_GetVolumeByID(t *testing.T) {
	ch := setup()
	expectedV := testVolumeCR
	err := ch.k8sClient.CreateCR(testCtx, expectedV.Name, &expectedV)
	assert.Nil(t, err)

	currentV, err := ch.GetVolumeByID(expectedV.Spec.Id)
	assert.Nil(t, err)
	assert.NotNil(t, currentV)
	assert.Equal(t, expectedV.Spec, currentV.Spec)

	// expected nil because of empty string as a ID
	volume, err := ch.GetVolumeByID("")
	assert.NotNil(t, err)
	assert.Nil(t, volume)
}

func TestCRHelper_GetDriveCRByUUID(t *testing.T) {
	ch := setup()
	expectedD := testDriveCR.DeepCopy()
	err := ch.k8sClient.CreateCR(testCtx, expectedD.Name, expectedD)
	assert.Nil(t, err)

	currentD := ch.GetDriveCRByUUID(expectedD.Spec.UUID)
	assert.NotNil(t, currentD)
	assert.Equal(t, expectedD.Spec, currentD.Spec)

	// expected nil because of empty string as a ID
	assert.Nil(t, ch.GetDriveCRByUUID(""))
}

func TestCRHelper_GetDriveCRByVolume(t *testing.T) {
	ch := setup()
	expectedV := testVolumeCR.DeepCopy()
	expectedV.Spec.Location = testLVGCR.Name
	expectedV.Spec.LocationType = v1.LocationTypeLVM
	err := ch.k8sClient.CreateCR(testCtx, expectedV.Name, expectedV)
	assert.Nil(t, err)
	// test LVG
	expectedLVG := testLVGCR.DeepCopy()
	expectedLVG.Spec.Locations = []string{testDriveCR.Name}
	err = ch.k8sClient.CreateCR(testCtx, expectedLVG.Name, expectedLVG)
	assert.Nil(t, err)
	testDriveCR1 := testDriveCR.DeepCopy()
	err = ch.k8sClient.CreateCR(testCtx, testDriveCR1.Name, testDriveCR1)
	assert.Nil(t, err)
	drive, err := ch.GetDriveCRByVolume(expectedV)
	assert.NotNil(t, drive)
	assert.Nil(t, err)
}

func TestCRHelper_GetVolumeCRs(t *testing.T) {
	ch := setup()
	v1 := testVolumeCR
	v2 := testVolumeCR
	v2.Name = "anotherName"
	v2.Spec.NodeId = "anotherNode"

	err := ch.k8sClient.CreateCR(testCtx, v1.Name, &v1)
	assert.Nil(t, err)
	err = ch.k8sClient.CreateCR(testCtx, v2.Name, &v2)
	assert.Nil(t, err)

	// node as empty string - expected all volumes
	currentVs, _ := ch.GetVolumeCRs()
	assert.NotNil(t, currentVs)
	assert.Equal(t, 2, len(currentVs))

	// expected one volume
	currentVs, _ = ch.GetVolumeCRs(v1.Spec.NodeId)
	assert.NotNil(t, currentVs)
	assert.Equal(t, 1, len(currentVs))
	assert.Equal(t, v1.Spec, currentVs[0].Spec)
}

func TestCRHelper_GetDriveCRs(t *testing.T) {
	ch := setup()
	d1 := testDriveCR.DeepCopy()
	d2 := testDriveCR.DeepCopy()
	d2.Name = "anotherName"
	d2.Spec.NodeId = "anotherNode"

	err := ch.k8sClient.CreateCR(testCtx, d1.Name, d1)
	assert.Nil(t, err)
	err = ch.k8sClient.CreateCR(testCtx, d2.Name, d2)
	assert.Nil(t, err)

	// node as empty string - expected all drives
	currentDs, _ := ch.GetDriveCRs()
	assert.NotNil(t, currentDs)
	assert.Equal(t, 2, len(currentDs))

	// expected one volume
	currentDs, _ = ch.GetDriveCRs(d1.Spec.NodeId)
	assert.NotNil(t, currentDs)
	assert.Equal(t, 1, len(currentDs))
	assert.Equal(t, d1.Spec, currentDs[0].Spec)
}

func TestCRHelper_GetVGNameByLVGCRName(t *testing.T) {
	ch := setup()
	lvgCR := testLVGCR
	err := ch.k8sClient.CreateCR(testCtx, lvgCR.Name, &lvgCR)
	assert.Nil(t, err)

	currentVGName, err := ch.GetVGNameByLVGCRName(lvgCR.Name)
	assert.Nil(t, err)
	assert.Equal(t, lvgCR.Spec.Name, currentVGName)

	// expected that LVG will not be found
	currentVGName, err = ch.GetVGNameByLVGCRName("randomName")
	assert.NotNil(t, err)
	assert.Equal(t, "", currentVGName)
}

// test AC deletion
func TestCRHelper_DeleteACsByNodeID(t *testing.T) {
	mock := setup()
	testACCRCopy := testACCR.DeepCopy()
	err := mock.k8sClient.CreateCR(testCtx, testACCR.Name, testACCRCopy)
	assert.Nil(t, err)

	err = mock.DeleteACsByNodeID(testACCRCopy.Spec.NodeId)
	assert.Nil(t, err)
}

// test Drive status update
func TestCRHelper_UpdateDrivesStatusOnNode(t *testing.T) {
	mock := setup()
	testDriveCRCopy := testDriveCR.DeepCopy()
	err := mock.k8sClient.CreateCR(testCtx, testDriveCRCopy.Name, testDriveCRCopy)
	assert.Nil(t, err)

	err = mock.UpdateDrivesStatusOnNode(testDriveCRCopy.Spec.NodeId, v1.DriveStatusOffline)
	assert.Nil(t, err)

	drive := mock.GetDriveCRByUUID(testDriveCRCopy.Name)
	assert.Equal(t, drive.Spec.Status, v1.DriveStatusOffline)
}

// test Volume operational status update
func TestCRHelper_UpdateVolumesOpStatusOnNode(t *testing.T) {
	mock := setup()
	err := mock.k8sClient.CreateCR(testCtx, testVolume.Name, testVolume.DeepCopy())
	assert.Nil(t, err)

	err = mock.UpdateVolumesOpStatusOnNode(testVolume.Spec.NodeId, v1.OperationalStatusMissing)
	assert.Nil(t, err)

	volume, err := mock.GetVolumeByID(testVolume.Name)
	assert.Nil(t, err)
	assert.Equal(t, volume.Spec.OperationalStatus, v1.OperationalStatusMissing)
}

func TestCRHelper_DeleteObjectByName(t *testing.T) {
	mock := setup()
	// object does not exist
	err := mock.DeleteObjectByName(testCtx, "aaaa", "", &accrd.AvailableCapacity{})
	assert.Nil(t, err)

	assert.Nil(t, mock.k8sClient.CreateCR(testCtx, testVolumeCR.Name, &testVolumeCR))
	assert.Nil(t, mock.DeleteObjectByName(testCtx, testVolumeCR.Name, "", &volumecrd.Volume{}))

	vList := &volumecrd.VolumeList{}
	assert.Nil(t, mock.k8sClient.ReadList(testCtx, vList))
	assert.Equal(t, 0, len(vList.Items))
}
