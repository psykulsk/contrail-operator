package templates

import "text/template"

// WebuiWebConfig is the template of the Webui Web service configuration.
var WebuiWebConfig = template.Must(template.New("").Parse(`/*
* Copyright (c) 2014 Juniper Networks, Inc. All rights reserved.
*/
var config = {};
config.orchestration = {};
config.orchestration.Manager = "{{ .Manager }}";
config.orchestrationModuleEndPointFromConfig = false;
config.contrailEndPointFromConfig = true;
config.regionsFromConfig = false;
config.endpoints = {};
config.endpoints.apiServiceType = "ApiServer";
config.endpoints.opServiceType = "OpServer";
config.regions = {};
{{- /* Create syntactically correct config when keystone not used. */}}
{{- if .KeystoneRegion }}
config.regions.{{ .KeystoneRegion }} = "{{ .KeystoneAuthProtocol }}://{{ .KeystoneAddress }}:{{ .KeystonePort }}/v3";
{{- else }}
config.regions.RegionOne = "{{ .KeystoneAuthProtocol }}://{{ .KeystoneAddress }}:{{ .KeystonePort }}/v3";
{{- end }}
config.serviceEndPointTakePublicURL = true;
config.networkManager = {};
config.networkManager.ip = "127.0.0.1";
config.networkManager.port = "9696";
config.networkManager.authProtocol = "http";
config.networkManager.apiVersion = [];
config.networkManager.strictSSL = false;
config.networkManager.ca = "";
config.imageManager = {};
config.imageManager.ip = "127.0.0.1";
config.imageManager.port = "9292";
config.imageManager.authProtocol = "http";
config.imageManager.apiVersion = ['v1', 'v2'];
config.imageManager.strictSSL = false;
config.imageManager.ca = "";
config.computeManager = {};
config.computeManager.ip = "127.0.0.1";
config.computeManager.port = "8774";
config.computeManager.authProtocol = "http";
config.computeManager.apiVersion = ['v1.1', 'v2'];
config.computeManager.strictSSL = false;
config.computeManager.ca = "";
config.identityManager = {};
config.identityManager.ip = "{{ .KeystoneAddress }}";
config.identityManager.port = "{{ .KeystonePort }}";
config.identityManager.authProtocol = "{{ .KeystoneAuthProtocol }}";
config.identityManager.apiVersion = ['v3'];
config.identityManager.strictSSL = false;
config.identityManager.defaultDomain = "{{ .KeystoneUserDomainName }}";
config.identityManager.ca = "{{ .CAFilePath }}";
config.storageManager = {};
config.storageManager.ip = "127.0.0.1";
config.storageManager.port = "8776";
config.storageManager.authProtocol = "http";
config.storageManager.apiVersion = ['v1'];
config.storageManager.strictSSL = false;
config.storageManager.ca = "";
config.cnfg = {};
config.cnfg.server_ip = [{{ .APIServerList }}];
config.cnfg.server_port = "{{ .APIServerPort }}";
config.cnfg.authProtocol = "https";
config.cnfg.strictSSL = false;
config.cnfg.ca = "{{ .CAFilePath }}";
config.cnfg.statusURL = '/global-system-configs';
config.analytics = {};
config.analytics.server_ip = [{{ .AnalyticsServerList }}];
config.analytics.server_port = "{{ .AnalyticsServerPort }}";
config.analytics.authProtocol = "https";
config.analytics.strictSSL = false;
config.analytics.ca = '{{ .CAFilePath }}';
config.analytics.statusURL = '/analytics/uves/bgp-peers';
config.dns = {};
config.dns.server_ip = [{{ .ControlNodeList }}];
config.dns.server_port = '{{ .DnsNodePort }}';
config.dns.statusURL = '/Snh_PageReq?x=AllEntries%20VdnsServersReq';
config.vcenter = {};
config.vcenter.server_ip = "127.0.0.1";         //vCenter IP
config.vcenter.server_port = "443";                                //Port
config.vcenter.authProtocol = "https";   //http or https
config.vcenter.datacenter = "vcenter";      //datacenter name
config.vcenter.dvsswitch = "vswitch";         //dvsswitch name
config.vcenter.strictSSL = false;                                  //Validate the certificate or ignore
config.vcenter.ca = '';                                            //specify the certificate key file
config.vcenter.wsdl = "/usr/src/contrail/contrail-web-core/webroot/js/vim.wsdl";
config.introspect = {};
config.introspect.ssl = {};
config.introspect.ssl.enabled = true;
config.introspect.ssl.key = '/etc/certificates/server-key-{{ .HostIP }}.pem';
config.introspect.ssl.cert = '/etc/certificates/server-{{ .HostIP }}.crt';
config.introspect.ssl.ca = '{{ .CAFilePath }}';
config.introspect.ssl.strictSSL = false;
config.jobServer = {};
config.jobServer.server_ip = '127.0.0.1';
config.jobServer.server_port = '3000';
config.files = {};
config.files.download_path = '/tmp';
config.cassandra = {};
config.cassandra.server_ips = [{{ .CassandraServerList }}];
config.cassandra.server_port = '{{ .CassandraPort }}';
config.cassandra.enable_edit = false;
config.cassandra.use_ssl = true;
config.cassandra.ca_certs = '{{ .CAFilePath }}';
config.kue = {};
config.kue.ui_port = '3002'
config.webui_addresses = {};
config.insecure_access = false;
config.http_port = '8180';
config.https_port = '8143';
config.require_auth = false;
config.node_worker_count = 1;
config.maxActiveJobs = 10;
config.redisDBIndex = 3;
config.CONTRAIL_SERVICE_RETRY_TIME = 300000; //5 minutes
config.redis_server_port = '{{ .RedisServerPort }}';
config.redis_server_ip = '{{ .RedisServerList }}';
config.redis_dump_file = '/var/lib/redis/dump-webui.rdb';
config.redis_password = '';
config.logo_file = '/opt/contrail/images/logo.png';
config.favicon_file = '/opt/contrail/images/favicon.ico';
config.featurePkg = {};
config.featurePkg.webController = {};
config.featurePkg.webController.path = '/usr/src/contrail/contrail-web-controller';
config.featurePkg.webController.enable = true;
config.qe = {};
config.qe.enable_stat_queries = false;
config.logs = {};
config.logs.level = 'debug';
config.getDomainProjectsFromApiServer = false;
config.network = {};
config.network.L2_enable = false;
config.getDomainsFromApiServer = false;
config.jsonSchemaPath = "/usr/src/contrail/contrail-web-core/src/serverroot/configJsonSchemas";
config.server_options = {};
config.server_options.key_file = '/etc/certificates/server-key-{{ .HostIP }}.pem';
config.server_options.cert_file = '/etc/certificates/server-{{ .HostIP }}.crt';
config.server_options.ciphers = 'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256:AES256-SHA';
module.exports = config;
config.staticAuth = [];
config.staticAuth[0] = {};
config.staticAuth[0].username = '{{ .AdminUsername }}';
config.staticAuth[0].password = '{{ .AdminPassword }}';
config.staticAuth[0].roles = ['cloudAdmin'];
`))

// WebuiAuthConfig is the template of the Webui Auth service configuration.
var WebuiAuthConfig = template.Must(template.New("").Parse(`/*
* Copyright (c) 2014 Juniper Networks, Inc. All rights reserved.
*/
var auth = {};
auth.admin_user = '{{ .AdminUsername }}';
auth.admin_password = '{{ .AdminPassword }}';
auth.admin_token = '';
auth.admin_tenant_name = '{{ .AdminUsername }}';
auth.project_domain_name = '{{ .KeystoneProjectDomainName }}';
auth.user_domain_name = '{{ .KeystoneUserDomainName }}';
module.exports = auth;
`))
