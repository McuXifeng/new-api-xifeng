import React from 'react';
import { Button, Space, Tag, Typography } from '@douyinfe/semi-ui';
import { timestamp2string } from '../../../helpers';
import TicketStatusTag from '../../ticket/TicketStatusTag';
import {
  getTicketPriorityColor,
  getTicketPriorityText,
  getTicketTypeText,
} from '../../ticket/ticketUtils';

const { Text } = Typography;

export const getTicketsColumns = ({
  t,
  admin = false,
  onOpenDetail,
  onCloseTicket,
}) => {
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      render: (value) => <Text strong>#{value}</Text>,
    },
    {
      title: t('主题'),
      dataIndex: 'subject',
      key: 'subject',
      render: (value, record) => (
        <div className='flex flex-col'>
          <Text strong>{value || '-'}</Text>
          <Text type='tertiary' size='small'>
            {getTicketTypeText(record?.type, t)}
          </Text>
        </div>
      ),
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (value) => <TicketStatusTag status={value} t={t} size='small' />,
    },
    {
      title: t('优先级'),
      dataIndex: 'priority',
      key: 'priority',
      width: 120,
      render: (value) => (
        <Tag color={getTicketPriorityColor(value)} shape='circle' size='small'>
          {getTicketPriorityText(value, t)}
        </Tag>
      ),
    },
  ];

  if (admin) {
    columns.push({
      title: t('用户'),
      dataIndex: 'username',
      key: 'username',
      width: 140,
      render: (value, record) => (
        <div className='flex flex-col'>
          <Text>{value || '-'}</Text>
          <Text type='tertiary' size='small'>
            UID: {record?.user_id || '-'}
          </Text>
        </div>
      ),
    });
  }

  columns.push(
    {
      title: t('更新时间'),
      dataIndex: 'updated_time',
      key: 'updated_time',
      width: 180,
      render: (value) => (value ? timestamp2string(value) : '-'),
    },
    {
      title: t('操作'),
      key: 'operate',
      width: 160,
      render: (_, record) => (
        <Space>
          <Button
            size='small'
            theme='borderless'
            type='primary'
            onClick={(event) => {
              event.stopPropagation();
              onOpenDetail?.(record);
            }}
          >
            {t('查看详情')}
          </Button>
          {!admin && Number(record?.status) !== 4 && (
            <Button
              size='small'
              theme='borderless'
              type='danger'
              onClick={(event) => {
                event.stopPropagation();
                onCloseTicket?.(record);
              }}
            >
              {t('关闭')}
            </Button>
          )}
        </Space>
      ),
    },
  );

  return columns;
};

