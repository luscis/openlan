# ping 
ansible openlan -m ping

# upgrade
ansible-playbook upgrade.yaml -e "version=0.8.22"
