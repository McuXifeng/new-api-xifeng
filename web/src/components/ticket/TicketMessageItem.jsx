import React from 'react';
import { Space, Tag, Typography } from '@douyinfe/semi-ui';
import { timestamp2string } from '../../helpers';

const { Text, Paragraph } = Typography;

const TicketMessageItem = ({ message, isMine, t }) => {
  const isAdmin = Number(message?.role || 0) >= 10;

  return (
    <div className={`flex ${isMine ? 'justify-end' : 'justify-start'}`}>
      <div
        className='w-full md:max-w-[78%] rounded-2xl px-4 py-3'
        style={{
          background: isMine
            ? 'var(--semi-color-primary-light-default)'
            : 'var(--semi-color-fill-0)',
          border: `1px solid ${
            isMine
              ? 'var(--semi-color-primary-light-hover)'
              : 'var(--semi-color-border)'
          }`,
        }}
      >
        <div className='flex items-start justify-between gap-3 mb-2'>
          <Space spacing={6} wrap>
            <Text strong>{message?.username || t('未知用户')}</Text>
            <Tag color={isAdmin ? 'orange' : 'blue'} shape='circle' size='small'>
              {isAdmin ? t('管理员') : t('用户')}
            </Tag>
          </Space>
          <Text type='tertiary' size='small'>
            {timestamp2string(message?.created_time || 0)}
          </Text>
        </div>
        <Paragraph
          className='!mb-0'
          style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}
        >
          {message?.content || '-'}
        </Paragraph>
      </div>
    </div>
  );
};

export default TicketMessageItem;

