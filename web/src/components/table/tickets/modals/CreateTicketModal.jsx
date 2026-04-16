import React, { useRef, useState } from 'react';
import {
  Form,
  Modal,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess } from '../../../../helpers';
import {
  getTicketPriorityOptions,
  getTicketTypeOptions,
} from '../../../ticket/ticketUtils';

const CreateTicketModal = ({ visible, onClose, onSuccess, t }) => {
  const [loading, setLoading] = useState(false);
  const formApiRef = useRef(null);

  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      const res = await API.post('/api/ticket/', {
        subject: values.subject,
        type: values.type,
        priority: values.priority,
        content: values.content,
      });
      if (res.data?.success) {
        showSuccess(t('工单创建成功'));
        onSuccess?.(res.data?.data);
        onClose?.();
      } else {
        showError(res.data?.message || t('工单创建失败'));
      }
    } catch (error) {
      showError(t('请求失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={t('创建工单')}
      visible={visible}
      onCancel={onClose}
      onOk={() => formApiRef.current?.submitForm()}
      okText={t('提交工单')}
      cancelText={t('取消')}
      confirmLoading={loading}
      centered
      width={560}
      style={{ maxWidth: '92vw' }}
      bodyStyle={{
        maxHeight: 'calc(80vh - 120px)',
        overflowY: 'auto',
        overflowX: 'hidden',
      }}
    >
      <Form
        layout='vertical'
        initValues={{
          type: 'general',
          priority: 2,
          subject: '',
          content: '',
        }}
        getFormApi={(api) => {
          formApiRef.current = api;
        }}
        onSubmit={handleSubmit}
      >
        <Form.Input
          field='subject'
          label={t('工单主题')}
          maxLength={255}
          showClear
          placeholder={t('请简要描述问题')}
          rules={[{ required: true, message: t('工单主题不能为空') }]}
        />
        <Form.Select
          field='type'
          label={t('工单类型')}
          optionList={getTicketTypeOptions(t)}
        />
        <Form.Select
          field='priority'
          label={t('优先级')}
          optionList={getTicketPriorityOptions(t)}
        />
        <Form.TextArea
          field='content'
          label={t('问题描述')}
          autosize={{ minRows: 4, maxRows: 8 }}
          maxLength={5000}
          showClear
          placeholder={t('请详细描述问题，方便管理员更快定位')}
          rules={[{ required: true, message: t('工单内容不能为空') }]}
        />
      </Form>
    </Modal>
  );
};

export default CreateTicketModal;

