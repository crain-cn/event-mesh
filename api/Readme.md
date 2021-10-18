 #steps
 go get -u github.com/swaggo/swag/cmd/swag
 swag init
 
 #doc 
 https://github.com/swaggo/swag/blob/master/README_zh-CN.md


#dev 
<pre>
sudo curl -L "https://github.com/docker/compose/releases/download/1.28.4/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
</pre>


<pre>
chmod 777 data/redis
docker-compose up -d
</pre>