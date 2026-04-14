import React, { useState } from 'react';
import { Button, Card, Space, TextArea, Typography } from '@douyinfe/semi-ui';
const { Title, Text } = Typography;

const TicketReplyBox = ({
  title,
  placeholder,
  submitText,
  disabled = false,
  loading = false,
  onSubmit,
  t,
}) => {
  const [content, setContent] = useState('');

  const handleSubmit = async () => {
    const nextContent = content.trim();
    if (!nextContent || disabled || loading) {
      return;
    }
    const ok = await onSubmit?.(nextContent);
    if (ok) {
      setContent('');
    }
  };

  return (
    <Card className='!rounded-2xl shadow-sm border-0'>
      <Space vertical align='start' style={{ width: '100%' }} spacing={12}>
        <div>
          <Title heading={5} className='!mb-1'>
            {title || t('回复工单')}
          </Title>
          {disabled && (
            <Text type='tertiary'>{t('当前工单已关闭，如需继续处理请先调整状态')}</Text>
          )}
        </div>
        <TextArea
          value={content}
          onChange={setContent}
          autosize={{ minRows: 4, maxRows: 8 }}
          maxLength={5000}
          showClear
          disabled={disabled || loading}
          placeholder={placeholder || t('请输入回复内容')}
        />
        <div className='w-full flex justify-end'>
          <Button
            theme='solid'
            type='primary'
            loading={loading}
            disabled={disabled || !content.trim()}
            onClick={handleSubmit}
          >
            {submitText || t('发送回复')}
          </Button>
        </div>
      </Space>
    </Card>
  );
};

export default TicketReplyBox;

