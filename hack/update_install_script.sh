#!/usr/bin/env bash

tmp_file="chaos-mesh.yaml"
tmp_install_scipt="install.sh.bak"
install_scirpt="install.sh"

helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --dry-run > ${tmp_file}

num=4
max=$(sed -n '$=' $tmp_file)
let sLine=max-num+1

sed -i .bak $sLine',$d' $tmp_file
sed -i .bak '1,9d' $tmp_file
sed -i .bak '/helm/d' $tmp_file
sed -i .bak '/Helm/d' $tmp_file
sed -i .bak 's/tls.crt:.*/tls.crt: \"\$\{TLS_CRT\}\"/g' $tmp_file
sed -i .bak 's/tls.key:.*/tls.key: \"\$\{TLS_KEY\}\"/g' $tmp_file
sed -i .bak 's/caBundle:.*/caBundle: \"\$\{CA_BUNDLE\}\"/g' $tmp_file
sed -i .bak 's/mountPath: \/var\/run\/docker.sock/mountPath: \$\{mountPath\}/g' $tmp_file
sed -i .bak 's/path: \/var\/run\/docker.sock/path: \$\{socketPath\}/g' $tmp_file
sed -i .bak 's/- docker/- $\{runtime\}/g' $tmp_file
mv $tmp_file $tmp_file.bak

cat <<EOF > $tmp_file
---
apiVersion: v1
kind: Namespace
metadata:
  name: chaos-testing
---
EOF

cat $tmp_file.bak >> $tmp_file

let start_num=$(cat -n $install_scirpt| grep "# chaos-mesh.yaml start" | awk '{print $1}')+1
let end_num=$(cat -n $install_scirpt| grep "# chaos-mesh.yaml end" | awk '{print $1}')-1

head -$start_num $install_scirpt > $tmp_install_scipt
cat $tmp_file >> $tmp_install_scipt
tail -n +$end_num $install_scirpt >> $tmp_install_scipt

mv $tmp_install_scipt $install_scirpt
chmod +x $install_scirpt

rm -rf $tmp_file
rm -rf $tmp_file.bak
rm -rf $tmp_install_scipt
