-- 1. 产品信息表（product_info）
CREATE TABLE product_info (
    product_id VARCHAR(32) PRIMARY KEY COMMENT '产品唯一标识',
    product_name VARCHAR(128) NOT NULL COMMENT '产品名称',
    product_type VARCHAR(32) NOT NULL COMMENT '产品类型',
    product_sub_type VARCHAR(32) COMMENT '产品子类型',
    risk_level VARCHAR(10) NOT NULL COMMENT '产品风险等级',
    product_status VARCHAR(20) NOT NULL DEFAULT '正常' COMMENT '产品状态',
    manager VARCHAR(64) NOT NULL COMMENT '基金管理人',
    custodian VARCHAR(64) NOT NULL COMMENT '基金托管人',
    purchase_fee_rate DECIMAL(10,4) COMMENT '前端申购费率',
    redemption_fee_rule JSON COMMENT '赎回费率规则',
    description TEXT COMMENT '产品说明书',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '产品创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '产品信息更新时间',
    INDEX idx_product_type (product_type),
    INDEX idx_risk_level (risk_level),
    INDEX idx_product_status (product_status),
    INDEX idx_manager (manager),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='产品信息表';

-- 2. 客户信息表（customer_info）
CREATE TABLE customer_info (
    customer_id VARCHAR(32) PRIMARY KEY COMMENT '客户唯一标识',
    customer_name VARCHAR(64) NOT NULL COMMENT '客户姓名/企业名称',
    customer_type VARCHAR(20) NOT NULL COMMENT '客户类型',
    id_type VARCHAR(20) NOT NULL COMMENT '证件类型',
    id_number VARCHAR(64) NOT NULL COMMENT '证件号码',
    risk_level VARCHAR(10) NOT NULL COMMENT '客户风险等级',
    risk_evaluation_time DATETIME NOT NULL COMMENT '风险测评时间',
    risk_evaluation_expire_time DATETIME NOT NULL COMMENT '风险测评过期时间',
    contact_phone VARCHAR(20) COMMENT '联系电话',
    email VARCHAR(64) COMMENT '电子邮箱',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '客户创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '客户信息更新时间',
    UNIQUE INDEX uk_id_number (id_number),
    INDEX idx_customer_type (customer_type),
    INDEX idx_risk_level (risk_level),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客户信息表';

-- 3. 产品净值表（product_net_value）
CREATE TABLE product_net_value (
    net_value_id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '净值记录唯一标识',
    product_id VARCHAR(32) NOT NULL COMMENT '产品标识',
    stat_date DATE NOT NULL COMMENT '净值统计日期',
    unit_net_value DECIMAL(10,4) NOT NULL COMMENT '单位净值',
    cumulative_net_value DECIMAL(10,4) NOT NULL COMMENT '累计净值',
    daily_growth_rate DECIMAL(10,4) COMMENT '日增长率',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    INDEX idx_product_id (product_id),
    INDEX idx_stat_date (stat_date),
    INDEX idx_product_stat (product_id, stat_date),
    UNIQUE INDEX uk_product_stat (product_id, stat_date) COMMENT '同一产品同一日期只能有一条净值记录',
    CONSTRAINT fk_net_value_product FOREIGN KEY (product_id) REFERENCES product_info(product_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='产品净值表';

-- 4. 客户银行卡表（customer_bank_card）
CREATE TABLE customer_bank_card (
    card_id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '银行卡记录唯一标识',
    customer_id VARCHAR(32) NOT NULL COMMENT '客户标识',
    bank_card_number VARCHAR(64) NOT NULL COMMENT '银行卡号，加密存储',
    bank_name VARCHAR(64) NOT NULL COMMENT '开户行名称',
    card_balance DECIMAL(18,2) NOT NULL DEFAULT 0.00 COMMENT '银行卡余额',
    is_virtual TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否虚拟银行卡，1-是，0-否',
    bind_status VARCHAR(20) NOT NULL DEFAULT '正常' COMMENT '绑定状态',
    bind_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
    unbind_time DATETIME COMMENT '解绑时间',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    INDEX idx_customer_id (customer_id),
    INDEX idx_bank_card_number (bank_card_number),
    INDEX idx_bind_status (bind_status),
    UNIQUE INDEX uk_card_customer (customer_id, bank_card_number) COMMENT '同一客户同一银行卡只能绑定一次',
    CONSTRAINT fk_bank_card_customer FOREIGN KEY (customer_id) REFERENCES customer_info(customer_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客户银行卡表';

-- 5. 申购申请表（purchase_application）
CREATE TABLE purchase_application (
    application_id VARCHAR(64) PRIMARY KEY COMMENT '申购申请编号',
    customer_id VARCHAR(32) NOT NULL COMMENT '客户标识',
    product_id VARCHAR(32) NOT NULL COMMENT '产品标识',
    card_id BIGINT NOT NULL COMMENT '银行卡标识',
    application_amount DECIMAL(18,2) NOT NULL COMMENT '申购金额',
    purchase_fee DECIMAL(18,2) COMMENT '申购费用',
    net_purchase_amount DECIMAL(18,2) COMMENT '净申购金额',
    application_date DATE NOT NULL COMMENT '申购申请日期（T日）',
    application_time DATETIME NOT NULL COMMENT '申购申请时间',
    expected_confirmation_date DATE NOT NULL COMMENT '预计确认日期（默认T+1）',
    application_status VARCHAR(20) NOT NULL DEFAULT '未确认' COMMENT '申请状态',
    risk_mismatch_remark VARCHAR(255) COMMENT '风险等级不匹配备注',
    operator_id VARCHAR(32) NOT NULL COMMENT '操作员ID',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    INDEX idx_customer_id (customer_id),
    INDEX idx_product_id (product_id),
    INDEX idx_card_id (card_id),
    INDEX idx_application_date (application_date),
    INDEX idx_expected_confirmation_date (expected_confirmation_date),
    INDEX idx_application_status (application_status),
    INDEX idx_operator_id (operator_id),
    INDEX idx_create_time (create_time),
    CONSTRAINT fk_purchase_customer FOREIGN KEY (customer_id) REFERENCES customer_info(customer_id),
    CONSTRAINT fk_purchase_product FOREIGN KEY (product_id) REFERENCES product_info(product_id),
    CONSTRAINT fk_purchase_bank_card FOREIGN KEY (card_id) REFERENCES customer_bank_card(card_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='申购申请表';

-- 6. 赎回申请表（redemption_application）
CREATE TABLE redemption_application (
    application_id VARCHAR(64) PRIMARY KEY COMMENT '赎回申请编号',
    customer_id VARCHAR(32) NOT NULL COMMENT '客户标识',
    product_id VARCHAR(32) NOT NULL COMMENT '产品标识',
    card_id BIGINT NOT NULL COMMENT '银行卡标识',
    application_shares DECIMAL(18,4) NOT NULL COMMENT '赎回份额',
    holding_period INT COMMENT '持有期限（天）',
    redemption_fee DECIMAL(18,2) COMMENT '赎回费用',
    expected_redemption_amount DECIMAL(18,2) COMMENT '预计赎回金额',
    application_date DATE NOT NULL COMMENT '赎回申请日期（T日）',
    application_time DATETIME NOT NULL COMMENT '赎回申请时间',
    expected_confirmation_date DATE NOT NULL COMMENT '预计确认日期（默认T+1）',
    application_status VARCHAR(20) NOT NULL DEFAULT '未确认' COMMENT '申请状态',
    operator_id VARCHAR(32) NOT NULL COMMENT '操作员ID',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    INDEX idx_customer_id (customer_id),
    INDEX idx_product_id (product_id),
    INDEX idx_card_id (card_id),
    INDEX idx_application_date (application_date),
    INDEX idx_expected_confirmation_date (expected_confirmation_date),
    INDEX idx_application_status (application_status),
    INDEX idx_operator_id (operator_id),
    INDEX idx_create_time (create_time),
    CONSTRAINT fk_redemption_customer FOREIGN KEY (customer_id) REFERENCES customer_info(customer_id),
    CONSTRAINT fk_redemption_product FOREIGN KEY (product_id) REFERENCES product_info(product_id),
    CONSTRAINT fk_redemption_bank_card FOREIGN KEY (card_id) REFERENCES customer_bank_card(card_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='赎回申请表';

-- 7. 客户持仓表（customer_position）
CREATE TABLE customer_position (
    position_id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '持仓记录唯一标识',
    customer_id VARCHAR(32) NOT NULL COMMENT '客户标识',
    product_id VARCHAR(32) NOT NULL COMMENT '产品标识',
    card_id BIGINT NOT NULL COMMENT '银行卡标识',
    total_shares DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '总持仓份额',
    available_shares DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '可用持仓份额',
    frozen_shares DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '冻结份额',
    average_cost DECIMAL(10,4) COMMENT '持仓平均成本',
    position_date DATE NOT NULL COMMENT '持仓统计日期',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '持仓创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '持仓更新时间',
    INDEX idx_customer_id (customer_id),
    INDEX idx_product_id (product_id),
    INDEX idx_card_id (card_id),
    INDEX idx_position_date (position_date),
    UNIQUE INDEX uk_customer_product_card (customer_id, product_id, card_id) COMMENT '同一客户同一产品同一银行卡只能有一条持仓记录',
    CONSTRAINT fk_position_customer FOREIGN KEY (customer_id) REFERENCES customer_info(customer_id),
    CONSTRAINT fk_position_product FOREIGN KEY (product_id) REFERENCES product_info(product_id),
    CONSTRAINT fk_position_bank_card FOREIGN KEY (card_id) REFERENCES customer_bank_card(card_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客户持仓表';

-- 8. 交易确认表（transaction_confirmation）
CREATE TABLE transaction_confirmation (
    confirmation_id VARCHAR(64) PRIMARY KEY COMMENT '确认编号',
    application_id VARCHAR(64) NOT NULL COMMENT '申请编号',
    transaction_type VARCHAR(20) NOT NULL COMMENT '交易类型',
    customer_id VARCHAR(32) NOT NULL COMMENT '客户标识',
    product_id VARCHAR(32) NOT NULL COMMENT '产品标识',
    confirmation_date DATE NOT NULL COMMENT '确认日期（T+1日）',
    confirmed_shares DECIMAL(18,4) COMMENT '确认份额（申购时有效）',
    confirmed_amount DECIMAL(18,2) COMMENT '确认金额（赎回时有效）',
    net_value DECIMAL(10,4) NOT NULL COMMENT '确认时使用的产品净值',
    fee DECIMAL(18,2) COMMENT '交易费用',
    confirmation_status VARCHAR(20) NOT NULL DEFAULT '确认成功' COMMENT '确认状态',
    failure_reason VARCHAR(255) COMMENT '确认失败原因',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '确认记录创建时间',
    INDEX idx_application_id (application_id),
    INDEX idx_transaction_type (transaction_type),
    INDEX idx_customer_id (customer_id),
    INDEX idx_product_id (product_id),
    INDEX idx_confirmation_date (confirmation_date),
    INDEX idx_confirmation_status (confirmation_status),
    INDEX idx_create_time (create_time),
    INDEX idx_app_type (application_id, transaction_type),
    CONSTRAINT fk_confirmation_customer FOREIGN KEY (customer_id) REFERENCES customer_info(customer_id),
    CONSTRAINT fk_confirmation_product FOREIGN KEY (product_id) REFERENCES product_info(product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易确认表';

-- 9. 清算日志表（liquidation_log）
CREATE TABLE liquidation_log (
    log_id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '日志记录唯一标识',
    liquidation_date DATE NOT NULL COMMENT '清算日期（T+1日）',
    liquidation_step VARCHAR(50) NOT NULL COMMENT '清算步骤',
    step_status VARCHAR(20) NOT NULL COMMENT '步骤状态',
    start_time DATETIME COMMENT '步骤开始时间',
    end_time DATETIME COMMENT '步骤结束时间',
    processed_count INT DEFAULT 0 COMMENT '处理记录数',
    failure_count INT DEFAULT 0 COMMENT '失败记录数',
    failure_detail TEXT COMMENT '失败详情',
    operator_id VARCHAR(32) COMMENT '操作员ID',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '日志创建时间',
    INDEX idx_liquidation_date (liquidation_date),
    INDEX idx_liquidation_step (liquidation_step),
    INDEX idx_step_status (step_status),
    INDEX idx_operator_id (operator_id),
    INDEX idx_create_time (create_time),
    INDEX idx_date_step (liquidation_date, liquidation_step)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='清算日志表';

-- 10. 系统角色表（sys_role）
CREATE TABLE sys_role (
    role_id VARCHAR(32) PRIMARY KEY COMMENT '角色唯一标识',
    role_name VARCHAR(64) NOT NULL COMMENT '角色名称',
    role_desc VARCHAR(255) COMMENT '角色描述',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '角色创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '角色信息更新时间',
    UNIQUE INDEX uk_role_name (role_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统角色表';

-- 11. 系统用户表（sys_user）
CREATE TABLE sys_user (
    user_id VARCHAR(32) PRIMARY KEY COMMENT '系统用户唯一标识',
    user_name VARCHAR(64) NOT NULL COMMENT '用户名',
    real_name VARCHAR(64) NOT NULL COMMENT '真实姓名',
    password VARCHAR(128) NOT NULL COMMENT '密码',
    role_id VARCHAR(32) NOT NULL COMMENT '角色ID',
    department VARCHAR(64) COMMENT '所属部门',
    position VARCHAR(64) COMMENT '职位',
    contact_phone VARCHAR(20) COMMENT '联系电话',
    user_status VARCHAR(20) NOT NULL DEFAULT '启用' COMMENT '用户状态',
    last_login_time DATETIME COMMENT '最后登录时间',
    password_expire_time DATETIME NOT NULL COMMENT '密码过期时间',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '用户创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '用户信息更新时间',
    UNIQUE INDEX uk_user_name (user_name),
    INDEX idx_role_id (role_id),
    INDEX idx_user_status (user_status),
    INDEX idx_department (department),
    CONSTRAINT fk_user_role FOREIGN KEY (role_id) REFERENCES sys_role(role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统用户表';

-- 12. 系统权限表（sys_permission）
CREATE TABLE sys_permission (
    permission_id VARCHAR(32) PRIMARY KEY COMMENT '权限唯一标识',
    permission_name VARCHAR(64) NOT NULL COMMENT '权限名称',
    permission_code VARCHAR(64) NOT NULL COMMENT '权限编码',
    permission_type VARCHAR(20) NOT NULL COMMENT '权限类型',
    parent_permission_id VARCHAR(32) COMMENT '父权限ID',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '权限创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '权限信息更新时间',
    UNIQUE INDEX uk_permission_code (permission_code),
    INDEX idx_parent_permission_id (parent_permission_id),
    INDEX idx_permission_type (permission_type),
    CONSTRAINT fk_permission_parent FOREIGN KEY (parent_permission_id) REFERENCES sys_permission(permission_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统权限表';

-- 13. 角色权限关联表（sys_role_permission）
CREATE TABLE sys_role_permission (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '关联记录唯一标识',
    role_id VARCHAR(32) NOT NULL COMMENT '角色ID',
    permission_id VARCHAR(32) NOT NULL COMMENT '权限ID',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '关联记录创建时间',
    UNIQUE INDEX uk_role_permission (role_id, permission_id),
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id),
    CONSTRAINT fk_role_permission_role FOREIGN KEY (role_id) REFERENCES sys_role(role_id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permission_permission FOREIGN KEY (permission_id) REFERENCES sys_permission(permission_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';

-- 14. 客户行为分析表（customer_behavior）
CREATE TABLE customer_behavior (
    behavior_id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '行为记录唯一标识',
    customer_id VARCHAR(32) NOT NULL COMMENT '客户标识',
    behavior_type VARCHAR(30) NOT NULL COMMENT '行为类型',
    behavior_time DATETIME NOT NULL COMMENT '行为发生时间',
    related_product_id VARCHAR(32) COMMENT '关联产品ID',
    behavior_detail JSON COMMENT '行为详情',
    ip_address VARCHAR(64) COMMENT 'IP地址',
    device_info VARCHAR(128) COMMENT '设备信息',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    INDEX idx_customer_id (customer_id),
    INDEX idx_behavior_type (behavior_type),
    INDEX idx_behavior_time (behavior_time),
    INDEX idx_related_product_id (related_product_id),
    INDEX idx_customer_behavior (customer_id, behavior_type),
    CONSTRAINT fk_behavior_customer FOREIGN KEY (customer_id) REFERENCES customer_info(customer_id),
    CONSTRAINT fk_behavior_product FOREIGN KEY (related_product_id) REFERENCES product_info(product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客户行为分析表';

-- 15. 工作日表（work_calendar）
CREATE TABLE work_calendar (
    calendar_date DATE PRIMARY KEY COMMENT '日期',
    is_workday TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否工作日，1-是，0-否',
    workday_type VARCHAR(20) NOT NULL DEFAULT '正常工作日' COMMENT '工作日类型',
    remark VARCHAR(255) COMMENT '备注',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    INDEX idx_is_workday (is_workday),
    INDEX idx_workday_type (workday_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作日表';

-- 插入客户信息数据
INSERT INTO customer_info (
    customer_id, customer_name, customer_type, id_type, id_number,
    risk_level, risk_evaluation_time, risk_evaluation_expire_time,
    contact_phone, email, create_time, update_time
) VALUES
-- 个人客户
('CUST2024000001', '张三', '个人', '身份证', '110101199001011234',
 'R3', '2024-01-15 09:30:00', '2025-01-15 09:30:00',
 '13800138001', 'zhangsan@example.com', '2024-01-10 10:00:00', '2024-05-20 14:30:00'),

('CUST2024000002', '李四', '个人', '身份证', '110101199002022345',
 'R4', '2024-02-20 14:15:00', '2025-02-20 14:15:00',
 '13800138002', 'lisi@example.com', '2024-02-15 11:20:00', '2024-05-21 10:45:00'),

('CUST2024000003', '王五', '个人', '身份证', '110101199003033456',
 'R2', '2024-03-10 10:00:00', '2025-03-10 10:00:00',
 '13800138003', 'wangwu@example.com', '2024-03-05 09:15:00', '2024-05-22 16:20:00'),

('CUST2024000004', '赵六', '个人', '身份证', '110101199004044567',
 'R5', '2024-04-05 15:30:00', '2025-04-05 15:30:00',
 '13800138004', 'zhaoliu@example.com', '2024-03-28 13:45:00', '2024-05-23 09:10:00'),

('CUST2024000005', '孙七', '个人', '身份证', '110101199005055678',
 'R3', '2024-05-12 11:45:00', '2025-05-12 11:45:00',
 '13800138005', 'sunqi@example.com', '2024-05-01 08:30:00', '2024-05-24 14:50:00'),

-- 企业客户
('CUST2024000006', '北京科技有限公司', '企业', '营业执照', '911101083456789012',
 'R4', '2024-01-20 16:00:00', '2025-01-20 16:00:00',
 '010-88888888', 'tech@beijingtech.com', '2024-01-15 14:20:00', '2024-05-25 11:30:00'),

('CUST2024000007', '上海贸易有限公司', '企业', '营业执照', '913101126543210987',
 'R3', '2024-02-25 10:30:00', '2025-02-25 10:30:00',
 '021-66666666', 'trade@shanghaitrade.com', '2024-02-18 09:45:00', '2024-05-26 15:40:00'),

('CUST2024000008', '广州实业集团有限公司', '企业', '营业执照', '91440101ABCD123456',
 'R5', '2024-03-30 13:15:00', '2025-03-30 13:15:00',
 '020-77777777', 'group@guangzhougroup.com', '2024-03-25 16:10:00', '2024-05-27 10:20:00');

-- 插入客户银行卡数据
INSERT INTO customer_bank_card (
    customer_id, bank_card_number, bank_name, card_balance,
    is_virtual, bind_status, bind_time, unbind_time, create_time, update_time
) VALUES
-- 张三的银行卡
('CUST2024000001', '6222021234567890123', '中国工商银行', 50000.00,
 0, '正常', '2024-01-10 10:30:00', NULL, '2024-01-10 10:30:00', '2024-05-20 14:35:00'),

('CUST2024000001', '6228481234567890124', '中国农业银行', 30000.00,
 0, '正常', '2024-02-15 11:00:00', NULL, '2024-02-15 11:00:00', '2024-05-21 09:15:00'),

-- 李四的银行卡
('CUST2024000002', '6227001234567890125', '中国建设银行', 80000.00,
 0, '正常', '2024-02-15 11:30:00', NULL, '2024-02-15 11:30:00', '2024-05-22 16:25:00'),

-- 王五的银行卡
('CUST2024000003', '6225881234567890126', '招商银行', 20000.00,
 0, '正常', '2024-03-05 09:45:00', NULL, '2024-03-05 09:45:00', '2024-05-23 09:15:00'),

('CUST2024000003', '6229091234567890127', '兴业银行', 15000.00,
 0, '正常', '2024-04-10 14:20:00', NULL, '2024-04-10 14:20:00', '2024-05-24 14:55:00'),

-- 赵六的银行卡
('CUST2024000004', '6222621234567890128', '交通银行', 100000.00,
 0, '正常', '2024-03-28 14:00:00', NULL, '2024-03-28 14:00:00', '2024-05-25 11:35:00'),

-- 孙七的银行卡
('CUST2024000005', '6226221234567890129', '中信银行', 40000.00,
 0, '正常', '2024-05-01 09:00:00', NULL, '2024-05-01 09:00:00', '2024-05-26 15:45:00'),

-- 虚拟银行卡（用于测试）
('CUST2024000005', 'V001202405200001', '虚拟银行', 100000.00,
 1, '正常', '2024-05-20 10:00:00', NULL, '2024-05-20 10:00:00', '2024-05-20 10:00:00'),

-- 北京科技有限公司的银行卡
('CUST2024000006', '6222081234567890130', '中国工商银行', 500000.00,
 0, '正常', '2024-01-15 15:00:00', NULL, '2024-01-15 15:00:00', '2024-05-27 10:25:00'),

-- 上海贸易有限公司的银行卡
('CUST2024000007', '6222021234567890131', '中国工商银行', 300000.00,
 0, '正常', '2024-02-18 10:30:00', NULL, '2024-02-18 10:30:00', '2024-05-28 09:30:00'),

-- 广州实业集团有限公司的银行卡
('CUST2024000008', '6227001234567890132', '中国建设银行', 1000000.00,
 0, '正常', '2024-03-25 16:45:00', NULL, '2024-03-25 16:45:00', '2024-05-29 14:40:00'),

('CUST2024000008', '6225881234567890133', '招商银行', 200000.00,
 0, '正常', '2024-04-05 10:15:00', NULL, '2024-04-05 10:15:00', '2024-05-30 16:20:00');