[![Docker Repository on Quay](https://quay.io/repository/omerlh/zaproxy-operator/status "Docker Repository on Quay")](https://quay.io/repository/omerlh/zaproxy-operator)
# zap-operator
A little operator that makes it easy to hack your existing applications in production. 
This tool is intended to run again your application, that you have permissions to attack.
Please do not use it for malicious purposes :)

[OWASP Zaproxy](https://www.owasp.org/index.php/OWASP_Zed_Attack_Proxy_Project) is a great security tool, that can be used to detect a lot of security tools.
This operator makes it easier to test your application in production.
To attack an application, all you need to do is:
* Install the operator (`helm repo add omerlh https://omerlh.github.io/zap-operator/ && helm install omerlh/zap-operator`)
* Create the CRD:
```
apiVersion: zaproxy.owasp.org/v1alpha1
kind: Zaproxy
metadata:
 name: example-zaproxy
spec:
 attackType: Passive
 tragetNamespace: default
 tragetIngress: <a name of exisitng ingress>
```
* Profit :)

The operator will create a new Zaproxy pod, and an Nginx Canary Ingress with 5% weight. 
All traffic passed to the canary ingress will be proxied by Zap.
Let it run for a while, you can always inspect Zap for alerts by running:
```
kubectl port-forward <zap pod name> 8090:8090
curl http://localhost:8090/OTHER/core/other/htmlreport/?formMethod=GET //get alerts in HTML format
```

## Known Limitations
* Only support Nginx Ingress
* Only support ingress with one host and one path
* Only support Ingress with backend service listening on port 80

## Roadmap
* Support Active attacks
* Support other ingress types
* Support service mesh (e.g. Istio/Linkerd)
* Publish to operator marketplace