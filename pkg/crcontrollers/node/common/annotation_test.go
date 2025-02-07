package common

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	fc "github.com/dell/csi-baremetal/pkg/base/featureconfig"
	"github.com/dell/csi-baremetal/pkg/base/k8s"
)

var (
	testNS     = ""
	testCtx    = context.Background()
	testLogger = logrus.New()

	nodeName = "node"
	nodeUID  = "11-22"
	testNode = coreV1.Node{
		ObjectMeta: metaV1.ObjectMeta{
			Name:        nodeName,
			UID:         types.UID(nodeUID),
			Namespace:   testNS,
			Annotations: map[string]string{},
		},
	}

	annotationKey   = "example/uuid"
	annotationValue = "1111-2222-3333-4444"
)

func TestGetNodeID(t *testing.T) {
	t.Run("All features disabled", func(t *testing.T) {
		featureConf := fc.NewFeatureConfig()

		nodeID, err := GetNodeID(&testNode, annotationKey, featureConf)
		assert.Equal(t, nodeUID, nodeID)
		assert.Nil(t, err)
	})

	t.Run("Default annotation feature", func(t *testing.T) {
		featureConf := fc.NewFeatureConfig()
		featureConf.Update(fc.FeatureNodeIDFromAnnotation, true)

		node := testNode.DeepCopy()
		node.Annotations[DeafultNodeIDAnnotationKey] = annotationValue

		nodeID, err := GetNodeID(node, annotationKey, featureConf)
		assert.Equal(t, annotationValue, nodeID)
		assert.Nil(t, err)
	})

	t.Run("Custom annotation feature", func(t *testing.T) {
		featureConf := fc.NewFeatureConfig()
		featureConf.Update(fc.FeatureNodeIDFromAnnotation, true)
		featureConf.Update(fc.FeatureExternalAnnotationForNode, true)

		node := testNode.DeepCopy()
		node.Annotations[annotationKey] = annotationValue

		nodeID, err := GetNodeID(node, annotationKey, featureConf)
		assert.Equal(t, annotationValue, nodeID)
		assert.Nil(t, err)
	})

	t.Run("Annotation is not set", func(t *testing.T) {
		featureConf := fc.NewFeatureConfig()
		featureConf.Update(fc.FeatureNodeIDFromAnnotation, true)
		featureConf.Update(fc.FeatureExternalAnnotationForNode, true)

		node := testNode.DeepCopy()

		_, err := GetNodeID(node, annotationKey, featureConf)
		assert.NotNil(t, err)
	})

	t.Run("Custom annotation feature is enabled, but annotationKey is empty", func(t *testing.T) {
		featureConf := fc.NewFeatureConfig()
		featureConf.Update(fc.FeatureNodeIDFromAnnotation, true)
		featureConf.Update(fc.FeatureExternalAnnotationForNode, true)

		node := testNode.DeepCopy()
		node.Annotations[annotationKey] = annotationValue

		_, err := GetNodeID(node, "", featureConf)
		assert.NotNil(t, err)
	})
}

func TestGetNodeIDByName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		k8sClient, err := k8s.GetFakeKubeClient(testNS, testLogger)
		assert.Nil(t, err)

		featureConf := fc.NewFeatureConfig()
		featureConf.Update(fc.FeatureNodeIDFromAnnotation, true)
		featureConf.Update(fc.FeatureExternalAnnotationForNode, true)

		node := testNode.DeepCopy()
		node.Annotations[annotationKey] = annotationValue
		assert.Nil(t, k8sClient.Create(testCtx, node))

		nodeID, err := GetNodeIDByName(k8sClient, nodeName, annotationKey, featureConf)
		assert.Equal(t, annotationValue, nodeID)
		assert.Nil(t, err)
	})

	t.Run("Node is not exist", func(t *testing.T) {
		k8sClient, err := k8s.GetFakeKubeClient(testNS, testLogger)
		assert.Nil(t, err)

		featureConf := fc.NewFeatureConfig()

		_, err = GetNodeIDByName(k8sClient, nodeName, annotationKey, featureConf)
		assert.NotNil(t, err)
	})
}
