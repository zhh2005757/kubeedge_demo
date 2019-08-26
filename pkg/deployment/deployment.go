package deployment

import (
    "github.com/gin-gonic/gin"
    "io/ioutil"
    appv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "github.com/json-iterator/go"
    "k8s.io/klog"
    kubeError "k8s.io/apimachinery/pkg/api/errors"
)

type Deployment struct{
    ClientSet *kubernetes.Clientset
}

func (n *Deployment) AddDeployment(ctx *gin.Context){
	body, _ := ioutil.ReadAll(ctx.Request.Body)
    deployment := &appv1.Deployment{}

    var json = jsoniter.ConfigCompatibleWithStandardLibrary
    json.Unmarshal(body, deployment)
    klog.Info(deployment)

    result, err := n.ClientSet.AppsV1().Deployments(deployment.GetNamespace()).Create(deployment)
    if err != nil {
        errRaw := err.(*kubeError.StatusError)
        ctx.JSON(int(errRaw.ErrStatus.Code), gin.H{
            "message": errRaw.ErrStatus.Status,
            "code":  errRaw.ErrStatus.Code,
            "reason":  errRaw.Error(),
        })
        return
    }

    ctx.JSON(200, result)
}

func (n *Deployment) DeleteDeployment(ctx *gin.Context){
	name := ctx.Param("name")
    nsName := ctx.Param("nsName")

    deployment, err := n.ClientSet.AppsV1().Deployments(nsName).Get(name, metav1.GetOptions{})
    if err != nil {
        errRaw := err.(*kubeError.StatusError)
        ctx.JSON(int(errRaw.ErrStatus.Code), gin.H{
            "message": errRaw.ErrStatus.Status,
            "code":  errRaw.ErrStatus.Code,
            "reason":  errRaw.Error(),
        })
        return
    }

    err = n.ClientSet.AppsV1().Deployments(nsName).Delete(name, &metav1.DeleteOptions{})
    if err != nil {
        errRaw := err.(*kubeError.StatusError)
        ctx.JSON(int(errRaw.ErrStatus.Code), gin.H{
            "message": errRaw.ErrStatus.Status,
            "code":  errRaw.ErrStatus.Code,
            "reason":  errRaw.Error(),
        })
        return
    }

   ctx.JSON(200,deployment)
}

func (n *Deployment) ListDeployment(ctx *gin.Context){
    nsName := ctx.Param("nsName")

    deploymentList, err := n.ClientSet.AppsV1().Deployments(nsName).List(metav1.ListOptions{})
    if err != nil {
        errRaw := err.(*kubeError.StatusError)
        ctx.JSON(int(errRaw.ErrStatus.Code), gin.H{
            "message": errRaw.ErrStatus.Status,
            "code":  errRaw.ErrStatus.Code,
            "reason":  errRaw.Error(),
        })
        return
    }
    ctx.JSON(200, deploymentList)
}

func (n *Deployment) GetDeployment(ctx *gin.Context){
	name := ctx.Param("name")
    nsName := ctx.Param("nsName")

    deployment, err := n.ClientSet.AppsV1().Deployments(nsName).Get(name, metav1.GetOptions{})
    if err != nil {
        errRaw := err.(*kubeError.StatusError)
        ctx.JSON(int(errRaw.ErrStatus.Code), gin.H{
            "message": errRaw.ErrStatus.Status,
            "code":  errRaw.ErrStatus.Code,
            "reason":  errRaw.Error(),
        })
        return
    }

    ctx.JSON(200, deployment)
}

func (n *Deployment) UpdateDeployment(ctx *gin.Context){
	body, _ := ioutil.ReadAll(ctx.Request.Body)
    deployment := &appv1.Deployment{}

    var json = jsoniter.ConfigCompatibleWithStandardLibrary
    json.Unmarshal(body, deployment)
    klog.Info(deployment)

    result, err := n.ClientSet.AppsV1().Deployments(deployment.GetNamespace()).Get(deployment.GetName(), metav1.GetOptions{})
    if err != nil {
        errRaw := err.(*kubeError.StatusError)
        ctx.JSON(int(errRaw.ErrStatus.Code), gin.H{
            "message": errRaw.ErrStatus.Status,
            "code":  errRaw.ErrStatus.Code,
            "reason":  errRaw.Error(),
        })
        return
    }

    result.SetLabels(deployment.GetLabels())
    result.SetAnnotations(deployment.GetAnnotations())
    result.Spec.Replicas = deployment.Spec.Replicas
    result.Spec.Template.Spec.Containers = deployment.Spec.Template.Spec.Containers

    updateResult, err := n.ClientSet.AppsV1().Deployments(deployment.GetNamespace()).Update(result)
    if err != nil {
        errRaw := err.(*kubeError.StatusError)
        ctx.JSON(int(errRaw.ErrStatus.Code), gin.H{
            "message": errRaw.ErrStatus.Status,
            "code":  errRaw.ErrStatus.Code,
            "reason":  errRaw.Error(),
        })
        return
    }
    
    ctx.JSON(200,updateResult)
}
