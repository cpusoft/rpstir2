package sys

// //////////////////////////////
// create
// //////////////////////////////
var createRpSqls []string = []string{
	`
#################################
## main table for cer/crl/roa/mft
#################################	
CREATE TABLE If Not Exists lab_rpki_cer (
	id int(10) unsigned not null primary key auto_increment,
	sn varchar(1024) NOT NULL,
	notBefore datetime NOT NULL,
	notAfter datetime NOT NULL,
	subject varchar(1024) ,
	issuer varchar(1024) ,
	ski varchar(128) ,
	aki varchar(128) ,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json not null comment 'all cer info in json',
	chainCerts json comment 'chain certs(cer/crl/mft/roa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key ski (ski),
	key aki (aki),
	key filePath (filePath(256)),
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId),
	unique cerFilePathFileName (filePath(256),fileName),
	unique cerSkiFilePath (ski,filePath(256))
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='main cer table'
`,

	`
CREATE TABLE If Not Exists lab_rpki_cer_sia (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	rpkiManifest varchar(512) comment 'mft sync url',
	rpkiNotify varchar(512),
	caRepository varchar(512) comment 'ca repository url(directory)',
	signedObject varchar(512) ,
	foreign key (cerid) REFERENCES lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer sia'
`,

	`
CREATE TABLE If Not Exists lab_rpki_cer_aia (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	caIssuers varchar(512) comment 'father ca url (cer file)',
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer aia'
`,

	`
CREATE TABLE If Not Exists lab_rpki_cer_crldp (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	crldp varchar(512) comment 'crl sync url(file)',
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer crl'
`,

	`
CREATE TABLE If Not Exists lab_rpki_cer_ipaddress (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	addressFamily int(10) unsigned not null,
	addressPrefix varchar(512) comment 'address prefix: 147.28.83.0/24 ',
	min varchar(512) comment 'min address: 99.96.0.0',
	max varchar(512) comment 'max address: 99.105.127.255',
	rangeStart varchar(512) comment 'min address range from addressPrefix or min/max, in hex:  63.60.00.00',
	rangeEnd varchar(512) comment 'max address range from addressPrefix or min/max, in hex:  63.69.7f.ff',
	addressPrefixRange json comment 'min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}',
	key addressPrefix (addressPrefix),
	key rangeStart (rangeStart),
	key rangeEnd (rangeEnd),
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer ip address range'
`,
	`
## because 0 of asn has special meaning, so the default of asn is -1, and is "bigint signed" in mysql
CREATE TABLE If Not Exists lab_rpki_cer_asn (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	asn bigint(20) signed,
	min bigint(20) signed,
	max bigint(20) signed,
	key asn (asn),
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer asn range'
`,

	`
###### crl
CREATE TABLE If Not Exists lab_rpki_crl (
	id int(10) unsigned not null primary key auto_increment,
	thisUpdate datetime NOT NULL,
	nextUpdate datetime NOT NULL,
	hasExpired varchar(8) ,
	aki varchar(128) ,
	crlNumber bigint(20) unsigned not null ,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json NOT NULL,
	chainCerts json comment 'chain certs(cer/crl/mft/roa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key aki (aki),
	key filePath (filePath(256)),
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId),
	unique crlFilePathFileName (filePath(256),fileName)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='crl '
`,

	`
CREATE TABLE If Not Exists lab_rpki_crl_revoked_cert (
	id int(10) unsigned not null primary key auto_increment,
	crlId int(10) unsigned not null,
	sn varchar(512) NOT NULL,
	revocationTime datetime NOT NULL,
	key sn (sn),
	foreign key (crlId) references lab_rpki_crl(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='all sn and revocationTime in crl'
`,

	`
###### manifest
CREATE TABLE If Not Exists lab_rpki_mft (
	id int(10) unsigned not null primary key auto_increment,
	mftNumber varchar(1024) NOT NULL,
	thisUpdate datetime NOT NULL,
	nextUpdate datetime NOT NULL,
	ski varchar(128) ,
	aki varchar(128) ,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json NOT NULL,
	chainCerts json comment 'chain certs(cer/crl/mft/roa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key ski (ski),
	key aki (aki),
	key filePath (filePath(256)),
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId),
	unique mftFilePathFileName (filePath(256),fileName),
	unique mftSkiFilePath (ski,filePath(256)) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='manifest'
`,

	`
CREATE TABLE If Not Exists lab_rpki_mft_sia (
	id int(10) unsigned not null primary key auto_increment,
	mftId int(10) unsigned not null,
	rpkiManifest varchar(512) ,
	rpkiNotify varchar(512) ,
	caRepository varchar(512) ,
	signedObject varchar(512) ,
	foreign key (mftId) REFERENCES lab_rpki_mft(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='mft sia'
`,

	`
CREATE TABLE If Not Exists lab_rpki_mft_aia (
	id int(10) unsigned not null primary key auto_increment,
	mftId int(10) unsigned not null,
	caIssuers varchar(512) ,
	foreign key (mftId) references lab_rpki_mft(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='mft aia'
`,

	`
CREATE TABLE If Not Exists lab_rpki_mft_file_hash (
	id int(10) unsigned not null primary key auto_increment,
	mftId int(10) unsigned not null,
	file varchar(1024),
	hash varchar(1024),
	key file(file(512)),
	foreign key (mftId) references lab_rpki_mft(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='files in manifest'
`,

	`
###### roa
## because 0 of asn has special meaning, so the default of asn is -1, and is "bigint signed" in mysql
CREATE TABLE If Not Exists lab_rpki_roa (
	id int(10) unsigned not null primary key auto_increment,
	asn bigint(20) signed not null,
	ski varchar(128) ,
	aki varchar(128) ,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json NOT NULL,
	chainCerts json comment 'chain certs(cer/crl/mft/roa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key ski (ski),
	key aki (aki),
	key filePath (filePath(256)),
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId),
	unique roaFilePathFileName (filePath(256),fileName)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa info'
`,
	//	lab_rpki_roa unique roaSkiFilePath (ski,filePath(256))
	`	
CREATE TABLE If Not Exists lab_rpki_roa_sia (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	rpkiManifest varchar(512) ,
	rpkiNotify varchar(512) ,
	caRepository varchar(512) ,
	signedObject varchar(512) ,
	foreign key (roaId) REFERENCES lab_rpki_roa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa sia'
`,

	`
CREATE TABLE If Not Exists lab_rpki_roa_aia (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	caIssuers varchar(512) ,
	foreign key (roaId) references lab_rpki_roa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa aia'
`,

	`
CREATE TABLE If Not Exists lab_rpki_roa_ipaddress (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	addressFamily int(10) unsigned not null,
	addressPrefix varchar(512),
	maxLength int(10) unsigned,
	rangeStart varchar(512),
	rangeEnd varchar(512),
	addressPrefixRange json comment 'min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}',
	key addressPrefix (addressPrefix),
	key rangeStart (rangeStart),
	key rangeEnd (rangeEnd),
	foreign key (roaId) references lab_rpki_roa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa ip prefix'
`,

	`
CREATE TABLE If Not Exists lab_rpki_roa_ee_ipaddress (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	addressFamily int(10) unsigned not null,
	addressPrefix varchar(512) comment 'address prefix: 147.28.83.0/24 ',
	min varchar(512) comment 'min address: 99.96.0.0',
	max varchar(512) comment 'max address: 99.105.127.255',
	rangeStart varchar(512) comment 'min address range from addressPrefix or min/max, in hex:  63.60.00.00',
	rangeEnd varchar(512) comment 'max address range from addressPrefix or min/max, in hex:  63.69.7f.ff',
	addressPrefixRange json comment 'min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}',
	key addressPrefix (addressPrefix),
	key rangeStart (rangeStart),
	key rangeEnd (rangeEnd),
	foreign key (roaId) references lab_rpki_roa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa ee ip prefix'
`,

	`
###### asa
CREATE TABLE If Not Exists lab_rpki_asa (
	id int(10) unsigned not null primary key auto_increment,
	ski varchar(128) ,
	aki varchar(128) ,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json NOT NULL,
	chainCerts json comment 'chain certs(cer/crl/mft/roa/asa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key ski (ski),
	key aki (aki),
	key filePath (filePath(256)),
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId),
	unique asaFilePathFileName (filePath(256),fileName),
	unique asaSkiFilePath (ski,filePath(256)) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='asa info'
`,

	`	
CREATE TABLE If Not Exists lab_rpki_asa_sia (
	id int(10) unsigned not null primary key auto_increment,
	asaId int(10) unsigned not null,
	rpkiManifest varchar(512) ,
	rpkiNotify varchar(512) ,
	caRepository varchar(512) ,
	signedObject varchar(512) ,
	foreign key (asaId) REFERENCES lab_rpki_asa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='asa sia'
`,

	`
CREATE TABLE If Not Exists lab_rpki_asa_aia (
	id int(10) unsigned not null primary key auto_increment,
	asaId int(10) unsigned not null,
	caIssuers varchar(512) ,
	foreign key (asaId) references lab_rpki_asa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='asa aia'
`,

	`
CREATE TABLE If Not Exists lab_rpki_asa_customer_provider_asn (
	id int(10) unsigned not null primary key auto_increment,
	asaId int(10) unsigned not null,
	customerAsn int(10) unsigned not null,
	providerAsn int(10) unsigned not null,
	providerAsnOrder int(10) unsigned not null,
	key customerAsn (customerAsn),
	key providerAsn (providerAsn),
	foreign key (asaId) references lab_rpki_asa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='asa customerAsn'
`,
	`
################################################
## recored every sync log for cer/crl/roa/mft/asa
################################################
CREATE TABLE If Not Exists lab_rpki_sync_log (
	id int(10) unsigned not null primary key auto_increment,
	syncState MEDIUMTEXT,
	parseValidateState MEDIUMTEXT,
	chainValidateState MEDIUMTEXT,
	rtrState MEDIUMTEXT,
	state text not null comment 'rsyncing/rsynced ddrping/ddrped  diffing/diffed   parsevalidating/parsevalidated   rtring/rtred idle',
	syncStyle varchar(255) not null comment 'rsync/rrdp' 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='recored every sync log'
`,

	`
CREATE TABLE If Not Exists lab_rpki_sync_log_file (
	id int(10) unsigned not null primary key auto_increment,
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncTime datetime not null comment 'sync time for every file',
	syncStyle varchar(16) not null comment 'rrdp/rsync',
	syncType varchar(16) not null comment 'add/del/update',
	fileType varchar(16) not null comment 'cer/roa/mft/crl/',
	filePath varchar(1024) NOT NULL,
	fileName varchar(128) NOT NULL,
	sourceUrl varchar(512),
	jsonAll json comment 'cert json info from cer/crl/mft/roa.jsonAll' ,
	fileHash varchar(512),
	state json comment '{"sync":"finished","updateCertTable":"notYet/finished"}: have synced ,have published to main table',
	key fileType (fileType),
	key syncType (syncType),
	key filePath (filePath(256)),
	key fileName (fileName),	
	foreign key (syncLogId) references lab_rpki_sync_log(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='recored sync log for cer/roa/mft/crl'
`,

	`
CREATE TABLE If Not Exists lab_rpki_sync_rrdp_log (
	id int(10) unsigned not null primary key auto_increment,
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	notifyUrl varchar(512) not null comment 'notification.xml url',
	sessionId varchar(512) not null comment 'session_id',
	lastSerial int(10) unsigned comment 'last serial',
	curSerial int(10) unsigned not null comment 'current serial',
	rrdpTime datetime not null comment 'rrdp time',
	rrdpType varchar(16) not null comment 'snapshot/delta' ,
	snapshotOrDeltaUrl varchar(256) not null comment 'snapshot/delta url' ,
	foreign key (syncLogId) references lab_rpki_sync_log(id),
	key notifyUrl (notifyUrl) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='recored notification.xml update log'
`,

	`
CREATE TABLE If Not Exists lab_rpki_sync_url (
	id int(10) unsigned not null primary key auto_increment,
	syncStyle varchar(16) not null comment 'rrdp/rsync',
	url varchar(256) not null comment 'rrdp/rsync url',
	rir varchar(256) comment 'rir',
	state json not null comment '{state:valid/invalid,****}',
	updateTime datetime comment 'rrdp update time',
	addTime datetime not null comment 'add time',
	unique syncUrl (url) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='all rsync/rrdp url'
`,
	`
CREATE TABLE If Not Exists lab_rpki_sync_rrdp_notify (
	id int(10) unsigned not null primary key auto_increment,
	preceptId varchar(16) not null comment 'same precept id',
	notifyUrl varchar(512) not null comment 'notification.xml url',
	version varchar(16) not null comment 'version',
	sessionId varchar(512) not null comment 'session_id',
	snapshotUrl varchar(512) not null comment 'snapshot url',
	snapshotHash varchar(512) not null comment 'snapshot hash',
	maxSerial int(10) unsigned comment 'max serial',
	minSerial int(10) unsigned comment 'min serial',
	curSerial int(10) unsigned comment 'current serial',
	state json comment '{state:valid}',
	preceptTime datetime not null comment 'precept time',
	downloadTime datetime comment 'download from notify time',
	unique notifyUrl (notifyUrl) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rrdp url'
`,
	`
CREATE TABLE If Not Exists lab_rpki_sync_rrdp_delta (
	id int(10) unsigned not null primary key auto_increment,
	preceptId varchar(16) not null comment 'same precept id',
	notifyUrl varchar(512) not null comment 'notification.xml url',
	deltaUrl  varchar(512) not null comment 'delta url',
	serial int(10) unsigned comment 'serial',
	deltaHash varchar(512) not null comment 'delta hash',
	state json comment '{state:valid}',
	updateTime datetime not null comment 'update time',
	key notifyUrl (notifyUrl),
	key deltaUrl (deltaUrl),
	key serial (serial)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rrdp url'
`,
	`
################################################
## sync distribute node
################################################
CREATE TABLE If Not Exists lab_rpki_distributed_node (
	id int(10) unsigned not null primary key auto_increment,
	nodeName varchar(256) not null comment 'node name, is urls host',
	nodeType varchar(16) not null comment 'center/node',
	url varchar(256) not null comment 'interface url: https://1.1.1.1:8080',
	note varchar(256) comment 'comment',
	state json not null comment '{state:valid/invalid/unknown}',
	performance json comment '{cpu:**,mem:**,disk:**,network:**,localtion:***}',
	location json comment '{}',
	repoStates json comment '[{}]',
	updateTime datetime not null comment 'update time',
	key nodeName(nodeName),
	key url(url)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='recored every sync log'
`,
	`
CREATE TABLE lab_rpki_distributed_select_detail  (
  id int(10) unsigned not null primary key auto_increment,
  logId varchar(100) not null,
  idx int not null,
  startTime datetime,
  selectTime datetime,
  detail json,
  key idxLogId(logId)
)  ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='detail'
`,

	`
##################
## RSS
##################
# 1 notify :: n delta_log
CREATE TABLE If Not Exists lab_rpki_rss_rrdp_notify (
	id int(10) unsigned not null primary key auto_increment,
	notifyUrl varchar(512) not null comment 'notification.xml url',
	sessionId varchar(512) not null comment 'session_id',
	latestSerial int(10) unsigned not null comment 'lateset delta serial',
	snapshotUrl varchar(512) not null comment 'snapshot url',
	updateTime datetime not null comment 'start update time',
	unique notifyUrl(notifyUrl)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rss rrdp notify'
`,
	`
CREATE TABLE If Not Exists lab_rpki_rss_rrdp_delta_log (	
	id int(10) unsigned not null primary key auto_increment,
	rssRrdpNotifyId int(10) unsigned not null comment 'rss rrdp notify id',
	deltaUrl varchar(512) not null comment 'delta url',
	serial int(10) unsigned not null comment 'serial',
	updateTime datetime not null comment 'start update time',
	saveTime datetime not null comment 'save time',
	deltaFiles mediumtext not null comment 'delta files:[{url:****,deltaStyle:publis/withdraw},{}]',
	key updateTime (updateTime),
	foreign key (rssRrdpNotifyId) references lab_rpki_rss_rrdp_notify(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rss rrdp delta log'
`,
	`
# 1 rsync :: n rsync_file_log
CREATE TABLE If Not Exists lab_rpki_rss_rsync (
	id int(10) unsigned not null primary key auto_increment,
	rsyncUrl varchar(512) not null comment 'rysnc url',
	updateTime datetime not null comment 'start update time',
	unique rsyncUrl(rsyncUrl)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rss rsync'
`,
	`
CREATE TABLE If Not Exists lab_rpki_rss_rsync_file_log (
	id int(10) unsigned not null primary key auto_increment,
	rssRsyncId int(10) unsigned not null comment 'rss rsync id',
	fileUrl varchar(512) not null comment 'file url',
	filePath varchar(1024) not null comment 'file path',
	fileName varchar(512) not null comment 'file name',
	fileHash varchar(512) not null comment 'file hash',
	style varchar(16) not null comment 'add/del/update',
	updateTime datetime not null comment 'start update time',
	saveTime datetime not null comment 'save time',
	key updateTime (updateTime),
	foreign key (rssRsyncId) references lab_rpki_rss_rsync(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rss rsync file log'
`,

	`
################################
## SLURM: init not drop
################################
CREATE TABLE If Not Exists lab_rpki_slurm (
	id int(10) unsigned not null primary key auto_increment,
	version int(10) unsigned default 1,
	style varchar(128) not null comment 'prefixFilter/bgpsecFilter/aspaFilter/hroaFilter/  prefixAssertion/bgpsecAssertion//aspaAssertion/hroaAssertion',

	asn int(10) unsigned comment 'prefix asn',
	addressPrefix varchar(512) comment 'prefix address: 198.51.100.0/24 or 2001:DB8::/32',
	maxLength int(10) unsigned  comment 'prefix maxlength',

	ski varchar(256) comment 'bgpsec base64 ski',
	routerPublicKey varchar(256) comment 'bgpsec base64 ski',

	customerAsn int(10) unsigned comment 'asa customerAsn',
	providerAsn  int(10) unsigned comment 'asa providerAsn',
	addressFamily varchar(16) comment 'asa addressFamily',

	hroaAsn  int(10) unsigned comment 'hroa Asn',
	subtreeIdentifier blob(128) comment 'hroa subtreeIdentifier',
	encodedSubtree int(10) unsigned comment 'hroa encodedSubtree',
	afiFlags  int(10) unsigned comment 'hroa afiFlags',

	customerAsnAsra int(10) unsigned comment 'asra customerAsnAsra',
	addressFamilyAsra varchar(16) comment 'asa addressFamily',
	providerAsnAsras json comment 'asra providerAsnAsras',
	otherNeighborAsnAsras json comment 'asra otherNeighborAsnAsras',
	customerAsnAsras json comment 'asra customerAsnAsras',
	lateralPeerAsnAsras json comment 'asra lateralPeerAsnAsras',
	hybridAsras json comment 'asra hybridAsras',
	valleyPathAsnAsras json comment 'asra valleyPathAsnAsras',

	comment varchar(256),
	slurmLogId int(10) unsigned not null comment 'lab_rpki_slurm_log.id',
	slurmLogFileId int(10) unsigned not null comment 'lab_rpki_slurm_log_file.id',
	state json not null comment '[rtr:notYet/finished]',

	key asn(asn),
	key addressPrefix(addressPrefix),
	key customerAsn(customerAsn),
	key providerAsn(providerAsn),
	key hroaAsn(hroaAsn),
	unique slurmPrefixAsaHroa (asn,addressPrefix,maxLength,customerAsn,providerAsn,addressFamily,hroaAsn,subtreeIdentifier(128),encodedSubtree)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='valid slurms'
`,

	`
### one slurm_log is corresponding to one slurm
CREATE TABLE If Not Exists lab_rpki_slurm_log (
	id int(10) unsigned not null primary key auto_increment,
	version int(10) unsigned default 1,
	style varchar(128) not null comment 'prefixFilter/bgpsecFilter/aspaFilter/hroaFilter   prefixAssertion/bgpsecAssertion/aspaAssertion/hroaAssertion/',

	asn int(10) unsigned comment 'prefix asn',
	addressPrefix varchar(512) comment 'prefix address: 198.51.100.0/24 or 2001:DB8::/32',
	maxLength int(10) unsigned  comment 'prefix maxlength',

	ski varchar(256) comment 'bgpsec base64 ski',
	routerPublicKey varchar(256) comment 'bgpsec base64 ski',

	customerAsn int(10) unsigned comment 'asa customerAsn',
	providerAsn  int(10) unsigned comment 'asa providerAsn',
	addressFamily varchar(16) comment 'asa addressFamily',

	hroaAsn  int(10) unsigned comment 'hroa Asn',
	subtreeIdentifier blob(128) comment 'hroa subtreeIdentifier',
	encodedSubtree int(10) unsigned comment 'hroa encodedSubtree',
	afiFlags  int(10) unsigned comment 'hroa afiFlags',


	customerAsnAsra int(10) unsigned comment 'asra customerAsnAsra',
	addressFamilyAsra varchar(16) comment 'asa addressFamily',
	providerAsnAsras json comment 'asra providerAsnAsras',
	otherNeighborAsnAsras json comment 'asra otherNeighborAsnAsras',
	customerAsnAsras json comment 'asra customerAsnAsras',
	lateralPeerAsnAsras json comment 'asra lateralPeerAsnAsras',
	hybridAsras json comment 'asra hybridAsras',
	valleyPathAsnAsras json comment 'asra valleyPathAsnAsras',

	comment varchar(256),
	state json not null comment '[state:unknown/valid/invalid]',
	slurmLogFileId int(10) unsigned not null comment 'lab_rpki_slurm_log_file.id',

	key asn(asn),
	key addressPrefix(addressPrefix),
	key customerAsn(customerAsn),
	key providerAsn(providerAsn)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='slurms from upload file, need audit'
`,

	`
### one slurm_log_file is corresponding to multi slurm_log and slurm
CREATE TABLE If Not Exists lab_rpki_slurm_log_file (
	id int(10) unsigned not null primary key auto_increment,
	content longtext not null comment 'slurm content',
	uploadUserId int(10) unsigned comment 'user upload slurm',
	uploadTime datetime NOT NULL,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='upload file'
`,
	`
### 1) one slurm_audit is corresponding to one slurm_log and one slurm
### 2) if one slurm pass, and then be unpass, there are will two slurm_audits
### and will delete slurm and slurm_log will be invalid
CREATE TABLE If Not Exists lab_rpki_slurm_audit (
	id int(10) unsigned not null primary key auto_increment,
	slurmId int(10) unsigned comment 'lab_rpki_slurm.id',
	slurmLogId int(10) unsigned not null comment 'lab_rpki_slurm_log.id',
	auditUserId int(10) unsigned comment 'user audit slurm',
	auditTime datetime comment 'audit time',
	state json not null comment '{state:unaudit/pass/unpass/del}',
	key slurmId(slurmId),
	key slurmLogId(slurmLogId)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='audit to slurm_log'
`,

	`
#####################
#### statistic
#####################
CREATE TABLE If Not Exists lab_rpki_statistic_rir (
	id int(10) unsigned not null primary key auto_increment,
	rir varchar(64) not null comment 'which nic',
	cerFileCount json not null comment 'cer Count',
	crlFileCount json not null comment 'crl Count',
	mftFileCount json not null comment 'mft Count',
	roaFileCount json not null comment 'roa Count',
	asaFileCount json not null comment 'asa Count',
	repos json not null comment 'repos, big json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)'
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='statis, update after every sync'
`,
	`
CREATE TABLE If Not Exists lab_rpki_statistic_count (
	id int(10) unsigned not null primary key auto_increment,
	cerCount int(10) not null comment 'cerCount',
	crlCount int(10) not null comment 'crlCount',
	mftCount int(10) not null comment 'mftCount',
	roaCount int(10) not null comment 'roaCount',
	asaCount int(10) not null comment 'asaCount',
	rtrCount int(10) not null comment 'rtrCount',
	slurmCount int(10) not null comment 'slurmCount',
	roaValidCount int(10) not null comment 'roaValidCount',
	roaWarningCount int(10) not null comment 'roaWarningCount',
	roaInvalidCount int(10) not null comment 'roaInvalidCount',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	countTime datetime not null comment 'count time'
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='statis, update after every sync'
`,
	`
CREATE TABLE If Not Exists lab_rpki_statistic_state (
	id int(10) unsigned not null primary key auto_increment,
	failInStateMsg varchar(1024) not null comment 'fail in stateMsg',
	fileNamesCount int(10) unsigned not null comment 'fileNamesCount',
	fileNames text not null comment 'file names',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)' 	
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='statis, update after every sync'
`,
	`
#####################
#### analyse
#####################
CREATE TABLE If Not Exists lab_rpki_analyse_roa_history (
	id int(10) unsigned not null primary key auto_increment,
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	roas json,
	updateTime datetime NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa history info'
`,
	/*
		   	`
		   ### python old version
		   #CREATE TABLE If Not Exists lab_rpki_analyse_roa_compete_py (
		   #	id int(10) unsigned not null primary key auto_increment,
		   #	fileName varchar(128) not null comment 'roa file name',
		   #	asn bigint(20) signed not null comment 'roa asn',
		   #	addressPrefixes json not null comment 'roa all prefix: [203.147.108.0/23,..,]',
		   #	competeResult json not null comment 'roa compete result, big json',
		   #	slurm json comment 'slurm',
		   #	updateTime datetime not null comment 'update time'
		   #) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa compete'

		`,
	*/
	` 
CREATE TABLE If Not Exists lab_rpki_analyse_roa_compete (
	id int(10) unsigned not null primary key auto_increment,
	curFileName varchar(128) not null comment 'cur roa file name',
	curAsn bigint(20) signed not null comment 'cur roa asn',
	curAddressPrefix varchar(128) not null comment 'cur compete prefix',
	curMaxLength bigint(20) not null comment 'cur compete maxlength',
	compFileName varchar(128) not null comment 'comp roa file name',
	compAsn bigint(20) signed not null comment 'comp roa asn',
	compAddressPrefix varchar(128) not null comment 'comp compete prefix',
	compMaxLength bigint(20) not null comment 'comp compete maxlength',
	competeDetail json not null comment 'comp roa compete detail json',
	updateTime datetime not null comment 'update time'
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa compete'  
`,
	` 
#####################
#### conf
#####################
CREATE TABLE If Not Exists lab_rpki_conf (
	id int(10) unsigned not null primary key auto_increment,
	section varchar(128) not null comment 'section',
	myKey varchar(128) not null comment 'key',
	myValue varchar(1024) not null comment 'value',
	defaultMyValue varchar(1024) not null comment 'default value',
	updateTime datetime not null comment 'update time',
	unique sectionMyKey (section,myKey)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rpstir2 configuration'
 

`,

	`
#########################
## create view roa ipaddress 
#########################
CREATE OR REPLACE VIEW lab_rpki_roa_ipaddress_view AS  
select r.id AS id,
	r.asn AS asn,
	i.addressPrefix AS addressPrefix,
	i.maxLength AS maxLength,
	r.syncLogId AS syncLogId,
	r.syncLogFileId AS syncLogFileId,
	r.origin->>'$.rir' as rir,
	r.origin->>'$.repo' as repo  
from lab_rpki_roa r join lab_rpki_roa_ipaddress i 
where i.roaId = r.id and 
	r.state->>'$.state' in ('valid','warning') 
order by 
	r.origin->>'$.rir',
	r.origin->>'$.repo',
	i.addressPrefix,
	i.maxLength,
	r.asn,
	r.id 
`,

	`
#########################
## create view crl revoked sn
#########################
CREATE OR REPLACE VIEW lab_rpki_crl_revoked_cert_view AS  
select l.id, l.fileName, l.aki, r.revocationTime, r.sn
from lab_rpki_crl l, lab_rpki_crl_revoked_cert r 
where l.id = r.crlId order by l.id
`,

	`
#########################
## create view mft file hash
#########################
CREATE OR REPLACE VIEW lab_rpki_mft_file_hash_view AS 
SELECT	m.id as mftId,	m.aki as aki,	fh.id as mftFileHashId,	fh.file as file, fh.hash as hash 
FROM lab_rpki_mft m, lab_rpki_mft_file_hash fh 
WHERE m.id = fh.mftId 
ORDER BY m.id, fh.id
`,

	`
#########################
## create view roaIpAddressCount
#########################
CREATE OR REPLACE VIEW lab_rpki_roa_ipaddress_count_view AS 
select roaId, count(*) as roaIpAddressCount 
from lab_rpki_roa_ipaddress 
group by roaId order by roaIpAddressCount 
`,
	`
#########################
## create view roaIpAddressCount
#########################
CREATE OR REPLACE VIEW lab_rpki_sync_rrdp_log_maxid_view AS 
select max(cc.id) AS maxId from lab_rpki_sync_rrdp_log cc group by cc.notifyUrl order by cc.notifyUrl 
`,

	`
#########################
## create view asa
#########################
CREATE OR REPLACE VIEW lab_rpki_asa_customer_provider_asns_view AS 
SELECT a.id, a.filepath,a.filename,cpa.asaId, cpa.customerAsn,GROUP_CONCAT(cpa.providerAsn) providerAsns 
FROM lab_rpki_asa_customer_provider_asn cpa 
	left join lab_rpki_asa a on a.id = cpa.asaId
GROUP BY cpa.asaId, cpa.customerAsn;
`,
}
var createVcSqls []string = []string{
	`
##################
## RTR
##################
CREATE TABLE If Not Exists lab_rpki_rtr_session (
	sessionId int(10) unsigned not null primary key comment 'sessionId, after init will not change',
	createTime datetime NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rtr session'
`,

	`
# insert random sessionId 
INSERT INTO lab_rpki_rtr_session(sessionId, createTime) VALUES(ROUND(RAND() * 999 + 99), NOW())
`,

	`
## serialNumber should not be auto_increment, because it will be wraped
CREATE TABLE If Not Exists lab_rpki_rtr_serial_number (
	id bigint(20) unsigned not null primary key auto_increment comment 'id',
	serialNumber bigint(20) unsigned not null comment 'serialNumber for rtr_full, rtr_incremental',
	globalSerialNumber bigint(20) unsigned not null comment 'serialNumber for center vc update by sync and slurm',
	subpartSerialNumber bigint(20) unsigned not null comment 'serialNumber for sub vc update by slurm',
	createTime datetime NOT NULL,
	unique rtrserialNumber (serialNumber) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='after every sync repo, serial num  will generate new serialnumber'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_full (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	asn bigint(20) signed not null,
	address varchar(512) not null comment 'address : 147.28.83 ',
	prefixLength int(10) unsigned not null,
	maxLength int(10) unsigned not null,
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key asn(asn),
	key address(address),
	key prefixLength(prefixLength),
	key maxLength(maxLength),
	unique rtrFullSerialNumberAsnAddressPrefixLengthMaxLength(serialNumber, asn, address, prefixLength, maxLength)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='after every sync repo, will insert all full'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_full_log (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	asn bigint(20) signed not null,
	address varchar(512) not null comment 'address : 147.28.83 ',
	prefixLength int(10) unsigned not null,
	maxLength int(10) unsigned not null,
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key asn(asn),
	key address(address),
	key prefixLength(prefixLength),
	key maxLength(maxLength) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr log history'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_incremental (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	style varchar(16) not null comment 'announce/withdraw, is 1/0 in protocol',
	asn bigint(20) signed not null,
	address varchar(512) not null comment 'address : 147.28.83 ',
	prefixLength int(10) unsigned not null,
	maxLength int(10) unsigned not null,
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key asn(asn),
	key address(address),
	key prefixLength(prefixLength),
	key maxLength(maxLength),
	unique rtrIncrementalSerialNumberAsnAddrPrefixMaxStyle(serialNumber, asn, address, prefixLength, maxLength, style)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='incremental rtr'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_asa_full (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	customerAsn int(10) unsigned not null comment 'customer asn',
	providerAsn int(10) unsigned not null comment 'provider asn',
	addressFamily int(10) unsigned,
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key customerAsn(customerAsn),
	key providerAsn(providerAsn),
	unique rtrAsaFull(serialNumber,customerAsn,providerAsn,addressFamily)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr asa'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_asa_full_log (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	customerAsn int(10) unsigned not null comment 'customer asn',
	providerAsn int(10) unsigned not null comment 'provider asn',
	addressFamily int(10) unsigned,
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key customerAsn(customerAsn),
	key providerAsn(providerAsn)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr asa log history'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_asa_incremental (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	style varchar(16) not null comment 'announce/withdraw, is 1/0 in protocol',
	customerAsn int(10) unsigned not null comment 'customer asn',
	providerAsn int(10) unsigned not null comment 'provider asn',
	addressFamily int(10) unsigned,
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key customerAsn(customerAsn),
	key providerAsn(providerAsn),
	unique rtrAsaIncremental(serialNumber,customerAsn,providerAsn,addressFamily)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='incremental rtr asa'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_hroa_full (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	
	hroaAsn  int(10) unsigned comment 'hroa Asn',
	subtreeIdentifier blob(128) comment 'hroa subtreeIdentifier',
	encodedSubtree int(10) unsigned comment 'hroa encodedSubtree',
	afiFlags int(10) unsigned comment 'hroa afi',
	
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber)	
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr asa'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_hroa_full_log (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	
	hroaAsn  int(10) unsigned comment 'hroa Asn',
	subtreeIdentifier blob(128) comment 'hroa subtreeIdentifier',
	encodedSubtree int(10) unsigned comment 'hroa encodedSubtree',
	afiFlags int(10) unsigned comment 'hroa afi',
	
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr asa log history'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_hroa_incremental (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	style varchar(16) not null comment 'announce/withdraw, is 1/0 in protocol',
	
	hroaAsn  int(10) unsigned comment 'hroa Asn',
	subtreeIdentifier blob(128) comment 'hroa subtreeIdentifier',
	encodedSubtree int(10) unsigned comment 'hroa encodedSubtree',
	afiFlags int(10) unsigned comment 'hroa afi',
	
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='incremental rtr asa'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_asra_full (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,

	customerAsnAsra int(10) unsigned comment 'asra customerAsnAsra',
	addressFamilyAsra varchar(16) comment 'asa addressFamily',
	providerAsnAsras json comment 'asra providerAsnAsras',
	otherNeighborAsnAsras json comment 'asra otherNeighborAsnAsras',
	customerAsnAsras json comment 'asra customerAsnAsras',
	lateralPeerAsnAsras json comment 'asra lateralPeerAsnAsras',
	hybridAsras json comment 'asra hybridAsras',
	valleyPathAsnAsras json comment 'asra valleyPathAsnAsras',

	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber)	
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr asra'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_asra_full_log (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,

	customerAsnAsra int(10) unsigned comment 'asra customerAsnAsra',
	addressFamilyAsra varchar(16) comment 'asa addressFamily',
	providerAsnAsras json comment 'asra providerAsnAsras',
	otherNeighborAsnAsras json comment 'asra otherNeighborAsnAsras',
	customerAsnAsras json comment 'asra customerAsnAsras',
	lateralPeerAsnAsras json comment 'asra lateralPeerAsnAsras',
	hybridAsras json comment 'asra hybridAsras',
	valleyPathAsnAsras json comment 'asra valleyPathAsnAsras',

	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr asra log history'
`,

	`
CREATE TABLE If Not Exists lab_rpki_rtr_asra_incremental (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	style varchar(16) not null comment 'announce/withdraw, is 1/0 in protocol',

	customerAsnAsra int(10) unsigned comment 'asra customerAsnAsra',
	addressFamilyAsra varchar(16) comment 'asa addressFamily',
	providerAsnAsras json comment 'asra providerAsnAsras',
	otherNeighborAsnAsras json comment 'asra otherNeighborAsnAsras',
	customerAsnAsras json comment 'asra customerAsnAsras',
	lateralPeerAsnAsras json comment 'asra lateralPeerAsnAsras',
	hybridAsras json comment 'asra hybridAsras',
	valleyPathAsnAsras json comment 'asra valleyPathAsnAsras',

	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='incremental rtr asra'
`,

	`
#####################
#### rush: init not drop
#####################
### one rush_node --> one rush_node_log --> many rush_node_audit
### create, insert rush_node_log/rush_node_audit,
### when pass, insert rush_node,update rush_node_log/rush_node_audit
### when unpass, update rush_node_log/rush_node_audit
### when del, del rush_node, update rush_node_log, insert rush_node_audit( set del)
	
CREATE TABLE If Not Exists lab_rpki_rush_node (
	id int(10) unsigned not null primary key auto_increment,
	nodeName varchar(256) not null comment 'node name',
	parentNodeId int(10) unsigned comment 'if it is root, will be null',
	url varchar(256) not null comment 'interface url: https://1.1.1.1:8080',
	isSelfUrl varchar(8) comment 'true/null: vc to identify itself. rp do not need this',
	note varchar(256) comment 'comments, copy from lab_rpki_rush_node_log.note, auditUser can change',
	updateTime datetime not null comment 'update time',
	unique nodeName(nodeName),
	unique url(url),
	key parentNodeId(parentNodeId) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rush node conf'
`,
	`
CREATE TABLE If Not Exists lab_rpki_rush_node_log (
	id int(10) unsigned not null primary key auto_increment,
	url varchar(256) not null comment 'interface url: https://1.1.1.1:8080',
	note varchar(256) comment 'comments',
	state json not null comment '{state:unknown/valid/invalid}',
	createTime datetime not null comment 'create time',
	createUserId int(10) not null comment 'create user',
	rushNodeId int(10) unsigned comment 'lab_rpki_rush_node id',
	key url(url)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rush node conf'
`,
	` 
CREATE TABLE If Not Exists lab_rpki_rush_node_audit (
	id int(10) unsigned not null primary key auto_increment,
	auditTime datetime comment 'audit time',
	auditUserId int(10) comment 'audit user',
	state json not null comment '{state:unaudit/pass/unpass/del}',
	rushNodeId int(10) unsigned comment 'lab_rpki_rush_node id,',
	rushNodeLogId int(10) unsigned not null comment 'lab_rpki_rush_node_log id',	
	key rushNodeId(rushNodeId),
	key rushNodeLogId(rushNodeLogId)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rush node conf'
`,
	`
CREATE TABLE If Not Exists lab_rpki_rush_transfer_log (
	id int(10) unsigned not null primary key auto_increment,
	uuid varchar(64) not null comment 'used to uniquely identify every rush transfer',
	sequence json not null comment 'identify transfer sequence, {seq:1, index:1}, to order',
	nodeName varchar(256) not null comment 'node name, to idengity node',
	nodeUrl varchar(256) not null comment 'node url: https://1.1.1.1:8080, to idengity node, not use nodeid',
	receiveRequestTime datetime comment 'receive request time, it is start time of process' ,
	sendResponseTime datetime comment 'send response time, it is end time of process' ,
	updateType varchar(64) not null comment 'requestfull/pushfull/pushincr',
	dataContent json comment 'some thing save, not use file to save',
	dataNumber int(10) unsigned comment 'the number of rpki data, will save in file',
	dataSha256 varchar(256) comment 'data sha256',
	filePath varchar(1024) comment 'saved file path',
	fileName varchar(256) comment 'saved file name',
	result varchar(16) comment 'ok/fail',
	errMsg varchar(256) comment 'fail reason',
	key uuid (uuid)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rush transfer log'
`,
}
var createLicenseSqls []string = []string{
	`
#####################
#### license: when init, not drop
#####################
CREATE TABLE If Not Exists lab_rpki_license_user (
	id int(10) unsigned not null primary key auto_increment,
	userName varchar(128) not null comment 'user name',
	state json not null comment '{state:valid/invalid}',
	licenseNumber int(10) unsigned not null comment 'allow license number',
	defaultDevicePeriod varchar(128) comment 'device period: 1Y,1M,1D',
	createTime datetime not null comment 'create time',
	unique userName (userName)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin COMMENT='user table'
`,

	`
CREATE TABLE If Not Exists lab_rpki_license_device (
	id int(10) unsigned not null primary key auto_increment,
	userId int(10) unsigned not null comment 'user id',
	keyId int(10) unsigned not null comment 'key id',
	deviceUuid varchar(128) not null comment 'device uuid, if info is empty, then use for encrypt',
	deviceName varchar(218) not null comment 'device name',

	rpVersion varchar(128) comment 'rp version',
	systemInfo varchar(512) comment 'system info',
  
	installTime datetime comment 'install time',
	licenseStartTime datetime comment 'start time',
	licenseEndTime datetime comment 'end time',
	state json not null comment '{state:valid/invalid}',

	unique deviceUuid (deviceUuid),
	unique userIdDeviceName(userId,deviceName)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin COMMENT='user table' 
`,

	`
CREATE TABLE If Not Exists lab_rpki_license_authorize_log (
	id int(10) unsigned not null primary key auto_increment,
	keyId int(10) unsigned not null comment 'key id',
	userName varchar(128) not null comment 'user name',
	deviceUuid varchar(128) null comment 'device uuid',
	
	decryptedInfo text comment 'decrypted info',
	authorizeTime  datetime not null comment 'create time',
	authorizeResult json not null comment '{verify:pass/unpass}'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin COMMENT='user table' 
`,

	`
CREATE TABLE If Not Exists lab_rpki_license_key (
	id int(10) unsigned not null primary key auto_increment,
	privateKey text not null comment 'private key',
	publicKey text not null comment 'public key',
	state varchar(64) not null comment '{state:valid/invalid}',
	createTime datetime not null comment 'create time'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin COMMENT='key table' 
`,
}

// //////////////////////////////
// drop
// //////////////////////////////
var dropRpSqls []string = []string{
	`drop table if exists lab_rpki_analyse_roa_compete`,
	`drop table if exists lab_rpki_analyse_roa_history`,
	`drop table if exists lab_rpki_conf`,
	`drop table if exists lab_rpki_cer_aia`,
	`drop table if exists lab_rpki_cer_asn`,
	`drop table if exists lab_rpki_cer_crldp`,
	`drop table if exists lab_rpki_cer_ipaddress`,
	`drop table if exists lab_rpki_cer_sia`,
	`drop table if exists lab_rpki_cer`,
	`drop table if exists lab_rpki_crl_revoked_cert`,
	`drop table if exists lab_rpki_crl`,
	`drop table if exists lab_rpki_mft_aia`,
	`drop table if exists lab_rpki_mft_file_hash`,
	`drop table if exists lab_rpki_mft_sia`,
	`drop table if exists lab_rpki_mft`,
	`drop table if exists lab_rpki_roa_aia`,
	`drop table if exists lab_rpki_roa_ee_ipaddress`,
	`drop table if exists lab_rpki_roa_ipaddress`,
	`drop table if exists lab_rpki_roa_sia`,
	`drop table if exists lab_rpki_roa`,
	`drop table if exists lab_rpki_asa_customer_provider_asn`,
	`drop table if exists lab_rpki_asa_aia`,
	`drop table if exists lab_rpki_asa_sia`,
	`drop table if exists lab_rpki_asa`,
	`drop table if exists lab_rpki_slurm`,
	`drop table if exists lab_rpki_slurm_log`,
	`drop table if exists lab_rpki_slurm_log_file`,
	`drop table if exists lab_rpki_slurm_audit`,
	`drop table if exists lab_rpki_rss_rrdp_delta_log`,
	`drop table if exists lab_rpki_rss_rrdp_notify`,
	`drop table if exists lab_rpki_rss_rsync_file_log`,
	`drop table if exists lab_rpki_rss_rsync`,
	`drop table if exists lab_rpki_statistic_rir`,
	`drop table if exists lab_rpki_statistic_count`,
	`drop table if exists lab_rpki_statistic_state`,
	`drop table if exists lab_rpki_sync_log_file`,
	`drop table if exists lab_rpki_sync_log`,
	`drop table if exists lab_rpki_sync_rrdp_log`,
	`drop table if exists lab_rpki_sync_url`,
	`drop table if exists lab_rpki_sync_rrdp_notify`,
	`drop table if exists lab_rpki_sync_rrdp_delta`,
	`drop table if exists lab_rpki_distributed_node`,
	`drop table if exists lab_rpki_distributed_select_detail`,
	`drop view if exists lab_rpki_crl_revoked_cert_view`,
	`drop view if exists lab_rpki_mft_file_hash_view`,
	`drop view if exists lab_rpki_roa_ipaddress_count_view`,
	`drop view if exists lab_rpki_roa_ipaddress_view`,
	`drop view if exists lab_rpki_sync_rrdp_log_maxid_view`,
	`drop view if exists lab_rpki_asa_customer_provider_asns_view`,
}

var dropVcSqls []string = []string{
	`drop table if exists lab_rpki_rtr_full_log`,
	`drop table if exists lab_rpki_rtr_full`,
	`drop table if exists lab_rpki_rtr_incremental`,
	`drop table if exists lab_rpki_rtr_asa_full_log`,
	`drop table if exists lab_rpki_rtr_asa_full`,
	`drop table if exists lab_rpki_rtr_asa_incremental`,
	`drop table if exists lab_rpki_rtr_hroa_full_log`,
	`drop table if exists lab_rpki_rtr_hroa_full`,
	`drop table if exists lab_rpki_rtr_hroa_incremental`,
	`drop table if exists lab_rpki_rtr_serial_number`,
	`drop table if exists lab_rpki_rtr_session`,
}
var dropLicenseSqls []string = []string{}

// //////////////////////////////
// truncate
// //////////////////////////////
var truncateRpSqls []string = []string{
	`truncate  table  lab_rpki_cer`,
	`truncate  table  lab_rpki_cer_sia`,
	`truncate  table  lab_rpki_cer_aia`,
	`truncate  table  lab_rpki_cer_crldp`,
	`truncate  table  lab_rpki_cer_ipaddress`,
	`truncate  table  lab_rpki_cer_asn`,
	`truncate  table  lab_rpki_crl`,
	`truncate  table  lab_rpki_crl_revoked_cert`,
	`truncate  table  lab_rpki_mft`,
	`truncate  table  lab_rpki_mft_sia`,
	`truncate  table  lab_rpki_mft_aia`,
	`truncate  table  lab_rpki_mft_file_hash`,
	`truncate  table  lab_rpki_roa`,
	`truncate  table  lab_rpki_roa_sia`,
	`truncate  table  lab_rpki_roa_aia`,
	`truncate  table  lab_rpki_roa_ipaddress`,
	`truncate  table  lab_rpki_roa_ee_ipaddress`,
	`truncate  table  lab_rpki_asa`,
	`truncate  table  lab_rpki_asa_sia`,
	`truncate  table  lab_rpki_asa_aia`,
	`truncate  table  lab_rpki_asa_customer_provider_asn`,
	`truncate  table  lab_rpki_slurm`,
	`truncate  table  lab_rpki_slurm_log`,
	`truncate  table  lab_rpki_slurm_log_file`,
	`truncate  table  lab_rpki_slurm_audit`,
	`truncate  table  lab_rpki_sync_rrdp_log`,
	`truncate  table  lab_rpki_sync_log_file`,
	`truncate  table  lab_rpki_sync_log`,
	`truncate  table  lab_rpki_sync_url`,
	`truncate  table  lab_rpki_sync_rrdp_notify`,
	`truncate  table  lab_rpki_sync_rrdp_delta`,
	`truncate  table  lab_rpki_distributed_node`,
	`truncate  table  lab_rpki_distributed_select_detail`,
	`truncate  table  lab_rpki_conf`,
	`truncate  table  lab_rpki_statistic_rir`,
	`truncate  table  lab_rpki_statistic_count`,
	`truncate  table  lab_rpki_statistic_state`,
	`truncate  table  lab_rpki_rss_rrdp_notify`,
	`truncate  table  lab_rpki_rss_rrdp_delta_log`,
	`truncate  table  lab_rpki_rss_rsync`,
	`truncate  table  lab_rpki_rss_rsync_file_log`,
	`truncate  table  lab_rpki_analyse_roa_history`,
	`truncate  table  lab_rpki_analyse_roa_compete`,
}
var truncateVcSqls []string = []string{
	`truncate  table  lab_rpki_rtr_session`,
	`truncate  table  lab_rpki_rtr_serial_number`,
	`truncate  table  lab_rpki_rtr_full`,
	`truncate  table  lab_rpki_rtr_full_log`,
	`truncate  table  lab_rpki_rtr_incremental`,
	`truncate  table  lab_rpki_rtr_asa_full`,
	`truncate  table  lab_rpki_rtr_asa_full_log`,
	`truncate  table  lab_rpki_rtr_asa_incremental`,
	`truncate  table  lab_rpki_rtr_hroa_full_log`,
	`truncate  table  lab_rpki_rtr_hroa_full`,
	`truncate  table  lab_rpki_rtr_hroa_incremental`,
}
var truncateLicenseSqls []string = []string{}

// //////////////////////////////
// truncate
// //////////////////////////////
var optimizeRpSqls []string = []string{
	`optimize  table  lab_rpki_cer`,
	`optimize  table  lab_rpki_cer_sia`,
	`optimize  table  lab_rpki_cer_aia`,
	`optimize  table  lab_rpki_cer_crldp`,
	`optimize  table  lab_rpki_cer_ipaddress`,
	`optimize  table  lab_rpki_cer_asn`,
	`optimize  table  lab_rpki_crl`,
	`optimize  table  lab_rpki_crl_revoked_cert`,
	`optimize  table  lab_rpki_mft`,
	`optimize  table  lab_rpki_mft_sia`,
	`optimize  table  lab_rpki_mft_aia`,
	`optimize  table  lab_rpki_mft_file_hash`,
	`optimize  table  lab_rpki_roa`,
	`optimize  table  lab_rpki_roa_sia`,
	`optimize  table  lab_rpki_roa_aia`,
	`optimize  table  lab_rpki_roa_ipaddress`,
	`optimize  table  lab_rpki_roa_ee_ipaddress`,
	`optimize  table  lab_rpki_asa`,
	`optimize  table  lab_rpki_asa_sia`,
	`optimize  table  lab_rpki_asa_aia`,
	`optimize  table  lab_rpki_asa_customer_provider_asn`,
	`optimize  table  lab_rpki_slurm`,
	`optimize  table  lab_rpki_slurm_log`,
	`optimize  table  lab_rpki_slurm_log_file`,
	`optimize  table  lab_rpki_slurm_audit`,
	`optimize  table  lab_rpki_sync_log_file`,
	`optimize  table  lab_rpki_sync_rrdp_log`,
	`optimize  table  lab_rpki_sync_log`,
	`optimize  table  lab_rpki_sync_url`,
	`optimize  table  lab_rpki_sync_rrdp_notify`,
	`optimize  table  lab_rpki_sync_rrdp_delta`,
	`optimize  table  lab_rpki_distributed_node`,
	`optimize  table  lab_rpki_distributed_select_detail`,
	`optimize  table  lab_rpki_rss_rrdp_notify`,
	`optimize  table  lab_rpki_rss_rrdp_delta_log`,
	`optimize  table  lab_rpki_rss_rsync`,
	`optimize  table  lab_rpki_rss_rsync_file_log`,
	`optimize  table  lab_rpki_statistic_rir`,
	`optimize  table  lab_rpki_statistic_count`,
	`optimize  table  lab_rpki_statistic_state`,
	`optimize  table  lab_rpki_analyse_roa_history`,
	`optimize  table  lab_rpki_analyse_roa_compete`,
}

var optimizeVcSqls []string = []string{
	`optimize  table  lab_rpki_rtr_session`,
	`optimize  table  lab_rpki_rtr_serial_number`,
	`optimize  table  lab_rpki_rtr_full`,
	`optimize  table  lab_rpki_rtr_full_log`,
	`optimize  table  lab_rpki_rtr_incremental`,
	`optimize  table  lab_rpki_rtr_asa_full`,
	`optimize  table  lab_rpki_rtr_asa_full_log`,
	`optimize  table  lab_rpki_rtr_asa_incremental`,
	`optimize  table  lab_rpki_rtr_hroa_full_log`,
	`optimize  table  lab_rpki_rtr_hroa_full`,
	`optimize  table  lab_rpki_rtr_hroa_incremental`,
	`optimize  table  lab_rpki_rush_node`,
	`optimize  table  lab_rpki_rush_node_audit`,
	`optimize  table  lab_rpki_rush_node_log`,
	`optimize  table  lab_rpki_rush_transfer_log`,
}

var optimizeLicenseSqls []string = []string{
	`optimize  table  lab_rpki_license_user`,
	`optimize  table  lab_rpki_license_device`,
	`optimize  table  lab_rpki_license_authorize_log`,
	`optimize  table  lab_rpki_license_key`,
}
