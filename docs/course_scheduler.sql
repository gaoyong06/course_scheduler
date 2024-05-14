CREATE TABLE `task` (
  `task_id` bigint(20) NOT NULL AUTO_INCREMENT '任务ID',
  `task_data` JSON NOT NULL '任务数据',
  `status` ENUM('pending', 'running', 'success', 'failed') NOT NULL '任务状态',
  `progress` tinyint(3) NOT NULL DEFAULT 0 COMMENT '任务进度(0-100)',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排课任务表';


CREATE TABLE `schedule_error_log` (
  `error_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '错误信息ID',
  `task_id` bigint(20) NOT NULL COMMENT '任务ID',
  `error_message` text NOT NULL COMMENT '错误信息',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`error_id`),
  KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排课错误表';


CREATE TABLE `schedule_result` (
  `result_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `task_id` bigint(20) NOT NULL COMMENT '任务ID',
  `subject_id` bigint(20) NOT NULL COMMENT '科目ID',
  `teacher_id` bigint(20) NOT NULL COMMENT '教师ID',
  `grade_id` bigint(20) NOT NULL COMMENT '班级ID',
  `class_id` bigint(20) NOT NULL COMMENT '班级ID',
  `venue_id` bigint(20) NOT NULL COMMENT '教学场地ID',
  `weekday` tinyint(3) NOT NULL COMMENT '周几',
  `period` tinyint(3) NOT NULL COMMENT '节次',
  `start_time` varchar(255) NOT NULL COMMENT '上课开始时间',
  `end_time` varchar(255) NOT NULL COMMENT '上课结束时间',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`result_id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_teacher_id` (`teacher_id`),
  KEY `idx_class_id` (`class_id`),
  KEY `idx_venue_id` (`venue_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排课结果表';

