package controller

import (
	"reflect"

	"github.com/appscode/glusterfs/api"
	"github.com/appscode/log"
	k8serror "k8s.io/kubernetes/pkg/api/errors"
)

func (c *Controller) update(old, new interface{}) {
	context := &options{}
	if oldGfs, ok := old.(*api.Glusterfs); ok {
		if newGfs, ok := new.(*api.Glusterfs); ok {
			if !reflect.DeepEqual(newGfs, oldGfs) {
				if c.validate(context, newGfs) {
					if !reflect.DeepEqual(oldGfs.Spec, newGfs.Spec) {
						restoreGfs := true
						if oldGfs.Spec.Replicas != newGfs.Spec.Replicas {
							c.updateStatefulSets(context, oldGfs)
							restoreGfs = false
						}

						if restoreGfs {
							gfs, err := c.ExtClient.Glusterfs(oldGfs.Namespace).Get(oldGfs.Name)
							if err != nil {
								log.Errorln("Error getting GlusterFS", err)
								return
							}
							gfs.Spec = oldGfs.Spec
						}
					}
				}
			}
		} else {
			log.Errorln("Failed to assert new Object to Glusterfs")
		}
	} else {
		log.Errorln("Failed to assert old Object to Glusterfs")
	}
}

func (c *Controller) updateStatefulSets(ctx *options, gfs *api.Glusterfs) error {
	ss, err := c.Client.Apps().StatefulSets(gfs.Namespace).Get(GlusterFSResourcePrefix + gfs.Name)
	if err != nil {
		log.Errorln("Error getting staetfulset", err)
		if k8serror.IsNotFound(err) {
			err := c.createStatefulSet(ctx, gfs)
			return err
		}
	}

	ss.Spec.Replicas = gfs.Spec.Replicas
	if ss.Spec.Replicas < 3 {
		ss.Spec.Replicas = 3
	}

	_, err = c.Client.Apps().StatefulSets(gfs.Namespace).Update(ss)
	if err != nil {
		log.Errorln("Error Updating staetfulset", err)
		return err
	}
	return nil
}
