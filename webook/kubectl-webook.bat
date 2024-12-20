set yamls=mysql-pv mysql-pvc mysql-service mysql-deployment redis-deployment redis-service deployment service ingress
(for %%y in (%yamls%) do (
    kubectl apply -f webook-%%y.yaml
))
