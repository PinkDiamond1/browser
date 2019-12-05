
CREATE TABLE IF NOT EXISTS block_id(
  id INT(11) NOT NULL AUTO_INCREMENT,
  hash VARCHAR(100) NOT NULL UNIQUE,
  parent_hash VARCHAR(100),
  height INT(11) NOT NULL UNIQUE,
  created INT(11) NOT NULL,
  gas_limit BIGINT NOT NULL,
  gas_used BIGINT NOT NULL ,
  fee VARCHAR(100),
  producer VARCHAR(100) NOT NULL,
  tx_count INT(11) NOT NULL,
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 10 * 1000000 每个块10笔交易，1000000块,按高度分表
CREATE TABLE IF NOT EXISTS block_tx_rel_id (
  id INT(11) NOT NULL AUTO_INCREMENT,
  height INT(11) NOT NULL,
  tx_hash VARCHAR(100),
  PRIMARY KEY (id),
  KEY `height_index` (`height`) USING BTREE
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS txs_hash (
  id INT(11) NOT NULL AUTO_INCREMENT,
  hash VARCHAR(100) NOT NULL UNIQUE,
  height INT(11) NOT NULL,
  gas_used BIGINT NOT NULL,
  gas_cost VARCHAR(100) NOT NULL,
  gas_price BIGINT NOT NULL,
  gas_asset_id INT(11) NOT NULL,
  state ENUM('0', '1') NOT NULL,
  block_hash VARCHAR(100) NOT NULL,
  tx_index INT(11) NOT NULL,
  action_count INT(11) NOT NULL,
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE IF NOT EXISTS actions_hash (
  id INT(11) NOT NULL AUTO_INCREMENT,
  tx_hash VARCHAR(100) NOT NULL,
  action_index INT(11) NOT NULL,
  action_hash VARCHAR(100) NOT NULL UNIQUE,
  nonce INT(11) NOT NULL,
  height INT(11) NOT NULL,
  created INT(11) NOT NULL,
  gas_asset_id INT(11) NOT NULL,
  transfer_asset_id INT(11) NOT NULL,
  action_type INT(11) NOT NULL,
  from_account VARCHAR(100) NOT NULL,
  to_account VARCHAR(100) NOT NULL,
  amount VARCHAR(1000) NOT NULL,
  gas_limit BIGINT NOT NULL,
  gas_used BIGINT NOT NULL,
  payer VARCHAR(100),
  payer_gas_price VARCHAR(1000),
  state ENUM('0', '1') NOT NULL,
  error_msg longtext,
  remark BLOB,
  payload BLOB,
  payload_size INT(11) NOT NULL,
  internal_action_count INT(11) NOT NULL,
  KEY `action_hash_index` (`action_hash`) USING BTREE,
  KEY `tx_hash_action_index` (`tx_hash`,`action_index`),
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE IF NOT EXISTS internal_actions_hash (
  id INT(11) NOT NULL AUTO_INCREMENT,
  tx_hash VARCHAR(100) NOT NULL,
  action_hash VARCHAR(100) NOT NULL,
  action_index INT(11) NOT NULL,
  internal_index INT(11) NOT NULL,
  height INT(11) NOT NULL,
  created INT(11) NOT NULL,
  asset_id INT(11) NOT NULL,
  action_type INT(11) NOT NULL,
  from_account VARCHAR(100) NOT NULL,
  to_account VARCHAR(100) NOT NULL,
  amount VARCHAR(1000) NOT NULL,
  gas_limit BIGINT NOT NULL,
  gas_used BIGINT NOT NULL,
  depth INT(11) NOT NULL,
  state ENUM('0', '1') NOT NULL,
  error_msg longtext,
  payload longtext,
  KEY `action_hash_index` (`action_hash`) USING BTREE,
  KEY `action_hash_internal_index` (`action_hash`,`internal_index`),
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS fee_actions_hash (
  id INT(11) NOT NULL AUTO_INCREMENT,
  tx_hash VARCHAR(100) NOT NULL,
  action_hash VARCHAR(100) NOT NULL,
  action_index INT(11) NOT NULL,
  fee_index INT(11) NOT NULL,
  height INT(11) NOT NULL,
  created INT(11) NOT NULL,
  asset_id INT(11) NOT NULL,
  from_account VARCHAR(100) NOT NULL,
  to_account VARCHAR(100) NOT NULL,
  amount VARCHAR(1000) NOT NULL,
  reason ENUM('0', '1', '2') NOT NULL,
  KEY `action_hash_index` (`action_hash`) USING BTREE,
  KEY `action_hash_internal_index` (`action_hash`,`fee_index`),
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS account (
  id INT(11) NOT NULL AUTO_INCREMENT,
  s_name VARCHAR(100) NOT NULL UNIQUE,
  parent_name VARCHAR(100),
  create_user VARCHAR(100),
  founder VARCHAR(100) NOT NULL,
  account_id INT(11) NOT NULL,
  account_number INT(11) NOT NULL,
  nonce INT(11) NOT NULL,
  author_version VARCHAR(100),
  threshold INT(11) NOT NULL,
  update_author_threshold INT(11) NOT NULL,
  permissions longtext NOT NULL,
  created INT(11) NOT NULL,
  contract_code longtext,
  code_hash VARCHAR(100),
  contract_created INT(11) DEFAULT 0,
  description BLOB,
  suicide INT(11) NOT NULL,
  destroy INT(11) NOT NULL,
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS balance_hash (
  id INT(11) NOT NULL AUTO_INCREMENT,
  account_name VARCHAR(100) NOT NULL,
  asset_id INT(11) NOT NULL,
  amount VARCHAR(1000) NOT NULL,
  update_height INT(11) NOT NULL,
  update_time INT(11) NOT NULL,
  KEY `account_name_asset_id` (`account_name`,`asset_id`),
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS token (
  id INT(11) NOT NULL AUTO_INCREMENT,
  asset_name VARCHAR(100) NOT NULL UNIQUE,
  asset_symbol VARCHAR(100) NOT NULL ,
  decimals INT(11) NOT NULL ,
  asset_id INT(11) NOT NULL UNIQUE ,
  contract_name VARCHAR(100) NOT NULL DEFAULT '',
  description BLOB,
  create_user VARCHAR(100) NOT NULL,
  create_time INT(11) NOT NULL,
  asset_owner VARCHAR(100) NOT NULL,
  founder VARCHAR(100) NOT NULL,
  upper_limit VARCHAR(1000) NOT NULL,
  liquidity VARCHAR(1000) NOT NULL,
  cumulative_issue VARCHAR(1000) NOT NULL,
  cumulative_destruction VARCHAR(1000) NOT NULL,
  update_time INT(11) NOT NULL,
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS account_action_history_hash (
  id INT(11) NOT NULL AUTO_INCREMENT,
  account_name VARCHAR(100) NOT NULL,
  tx_hash VARCHAR(100) NOT NULL,
  action_hash VARCHAR(100) NOT NULL,
  action_index INT(11) NOT NULL,
  other_index INT(11),
  tx_type ENUM('0', '1', '2', '3', '4', '5', '6', '7', '8') NOT NULL COMMENT '0外部转账，1内部转账，2外部合约调用，3内部合约调用，4外部合约被调，5内部合约被调，6外部其他交易，7内部其他交易，8手续费交易',
  height INT(11) NOT NULL,
  PRIMARY KEY (id),
  KEY `account_name_tx_type` (`account_name`,`tx_type`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS token_history (
  id INT(11) NOT NULL AUTO_INCREMENT,
  token_id INT(11) NOT NULL,
  tx_hash VARCHAR(100) NOT NULL,
  action_index INT(11) NOT NULL,
  action_hash VARCHAR(100) NOT NULL,
  internal_index INT(11) NOT NULL,
  tx_type INT(11) NOT NULL comment '0外部交易，1内部交易',
  action_type INT(11) NOT NULL,
  height INT(11) NOT NULL,
  KEY `token_id` (`token_id`),
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE IF NOT EXISTS token_fee_history_id (
  id INT(11) NOT NULL AUTO_INCREMENT,
  token_id int(11) NOT NULL,
  tx_hash VARCHAR(100) NOT NULL,
  action_index int(11) NOT NULL,
  action_hash VARCHAR(100) NOT NULL,
  fee_action_index int(11) NOT NULL,
  height INT(11) NOT NULL,
  KEY `token_id_height` (`token_id`,`height`),
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS task_status (
  id INT(11) NOT NULL AUTO_INCREMENT,
  task_type enum('block', 'txs', 'action', 'internalAction', 'feeAction', 'account', 'accountBalance', 'token', 'accountHistory', 'tokenHistory', 'feeHistory'),
  height INT(11) NOT NULL,
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS chain_status (
  id INT(11) NOT NULL AUTO_INCREMENT,
  height INT(11) NOT NULL,
  tx_count INT(11) NOT NULL,
  producer_number INT(11) NOT NULL,
  fee_income VARCHAR(1000) NOT NULL,
  token_income VARCHAR(1000) NOT NULL,
  contract_income VARCHAR(1000) NOT NULL,
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS block_rollback (
  id INT(11) NOT NULL AUTO_INCREMENT,
  block_data LONGBLOB NOT NULL,
  height INT(11) NOT NULL,
  block_hash VARCHAR(100) NOT NULL,
  parent_hash VARCHAR(100) NOT NULL,
  KEY `key_height` (`height`),
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS account_rollback (
  id INT(11) NOT NULL AUTO_INCREMENT,
  s_name VARCHAR(100),
  parent_name VARCHAR(100),
  create_user VARCHAR(100),
  founder VARCHAR(100) NOT NULL,
  account_id INT(11) NOT NULL,
  account_number INT(11) NOT NULL,
  nonce INT(11) NOT NULL,
  author_version VARCHAR(100),
  threshold INT(11) NOT NULL,
  update_author_threshold INT(11) NOT NULL,
  permissions longtext NOT NULL,
  created INT(11) NOT NULL,
  contract_code longtext,
  code_hash VARCHAR(100),
  contract_created INT(11) DEFAULT 0,
  description BLOB,
  suicide INT(11) NOT NULL,
  destroy INT(11) NOT NULL,
  height INT(11) NOT NULL,
  KEY `height_asset_name` (`height`,`s_name`),
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS token_rollback (
  id INT(11) NOT NULL AUTO_INCREMENT,
  height INT(11) NOT NULL,
  asset_name VARCHAR(100) NOT NULL,
  asset_symbol VARCHAR(100) NOT NULL ,
  decimals INT(11) NOT NULL ,
  asset_id INT(11) NOT NULL ,
  contract_name VARCHAR(100) NOT NULL DEFAULT '',
  description BLOB,
  create_user VARCHAR(100) NOT NULL,
  create_time INT(11) NOT NULL,
  asset_owner VARCHAR(100) NOT NULL,
  founder VARCHAR(100) NOT NULL,
  upper_limit VARCHAR(1000) NOT NULL,
  liquidity VARCHAR(1000) NOT NULL,
  cumulative_issue VARCHAR(1000) NOT NULL,
  cumulative_destruction VARCHAR(1000) NOT NULL,
  update_time INT(11) NOT NULL,
  KEY `height_asset_name` (`height`,`asset_name`),
  PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS statis_token (
  token_name    VARCHAR(100) NOT NULL,
  user_num      INT(11) NOT NULL,
  user_rank     INT(11) NOT NULL,
  call_num      INT(11) NOT NULL,
  call_rank     INT(11) NOT NULL,
  holder_num    INT(11) NOT NULL,
  holder_rank   INT(11) NOT NULL,
  income_rank   INT(11) NOT NULL,
  feeTotal      VARCHAR(100) NOT NULL,
  PRIMARY KEY (token_name)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- statis table 
CREATE TABLE IF NOT EXISTS statis_contract (
  contract_name VARCHAR(100) NOT NULL,
  user_num      INT(11) NOT NULL,
  user_rank     INT(11) NOT NULL,
  call_num      INT(11) NOT NULL,
  call_rank     INT(11) NOT NULL,
  income_rank   INT(11) NOT NULL,
  feeTotal      VARCHAR(100) NOT NULL,
  PRIMARY KEY (contract_name)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS statis_token_info (
  id       INT(11) NOT NULL AUTO_INCREMENT,
  name     VARCHAR(100) NOT NULL UNIQUE,
  decimals INT(11) NOT NULL,
  assetid  BIGINT NOT NULL UNIQUE,
  shortname VARCHAR(100) NOT NULL UNIQUE,
  PRIMARY  KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS statis_fee_total (
  name              VARCHAR(100) NOT NULL,
  nametype          ENUM('0', '1') NOT NULL COMMENT '0 token 1 contract',
  rank              INT(11) NOT NULL,
  fee               VARCHAR(100) NOT NULL,
  KEY `index_rank` (`rank`),
  PRIMARY KEY (name)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS statis_block_info (
  id INT(11) NOT NULL,
  height     BIGINT NOT NULL,
  PRIMARY    KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;