-- Casbin Rule Table
CREATE TABLE casbin_rule (
    id    BIGINT AUTO_INCREMENT PRIMARY KEY,
    ptype VARCHAR(255) DEFAULT '' NOT NULL,
    v0    VARCHAR(255) DEFAULT '' NOT NULL,
    v1    VARCHAR(255) DEFAULT '' NOT NULL,
    v2    VARCHAR(255) DEFAULT '' NOT NULL,
    v3    VARCHAR(255) DEFAULT '' NOT NULL,
    v4    VARCHAR(255) DEFAULT '' NOT NULL,
    v5    VARCHAR(255) DEFAULT '' NOT NULL,
    INDEX idx_ptype (ptype),
    INDEX idx_v0 (v0),
    INDEX idx_v1 (v1),
    INDEX idx_v2 (v2),
    INDEX idx_v3 (v3),
    INDEX idx_v4 (v4),
    INDEX idx_v5 (v5),
    UNIQUE INDEX uniq_ptype_v0_v1_v2_v3_v4_v5 (ptype, v0, v1, v2, v3, v4, v5)
) COMMENT 'Casbin';