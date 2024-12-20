set list=deployment service pod ingress pv pvc
(for %%a in (%list%) do (
    kubectl delete %%a --all
))