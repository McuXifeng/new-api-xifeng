import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Modal, Select, Space } from '@douyinfe/semi-ui';
import { IconPlusCircle } from '@douyinfe/semi-icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { useTableCompactMode } from '../../hooks/common/useTableCompactMode';
import TicketsPage from '../../components/table/tickets';
import CreateTicketModal from '../../components/table/tickets/modals/CreateTicketModal';
import CreateInvoiceTicketModal from '../../components/table/tickets/modals/CreateInvoiceTicketModal';
import {
  getTicketStatusOptions,
  getTicketTypeOptions,
} from '../../components/ticket/ticketUtils';

const Ticket = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [compactMode, setCompactMode] = useTableCompactMode('tickets-user');
  const [tickets, setTickets] = useState([]);
  const [ticketCount, setTicketCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showInvoiceModal, setShowInvoiceModal] = useState(false);

  const loadTickets = useCallback(async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/ticket/self', {
        params: {
          p: activePage,
          page_size: pageSize,
          status: statusFilter || undefined,
          type: typeFilter || undefined,
        },
      });
      if (res.data?.success) {
        const pageData = res.data?.data || {};
        setTickets(pageData.items || []);
        setTicketCount(Number(pageData.total || 0));
      } else {
        showError(res.data?.message || t('工单加载失败'));
      }
    } catch (error) {
      showError(t('请求失败'));
    } finally {
      setLoading(false);
    }
  }, [activePage, pageSize, statusFilter, typeFilter, t]);

  useEffect(() => {
    loadTickets();
  }, [loadTickets]);

  const handleCloseTicket = (ticket) => {
    Modal.confirm({
      title: t('确认关闭工单'),
      content: t('关闭后仍可查看历史消息，如需继续处理可由管理员重新调整状态。'),
      centered: true,
      onOk: async () => {
        try {
          const res = await API.put(`/api/ticket/self/${ticket.id}/close`);
          if (res.data?.success) {
            showSuccess(t('工单已关闭'));
            loadTickets();
          } else {
            showError(res.data?.message || t('关闭工单失败'));
          }
        } catch (error) {
          showError(t('请求失败'));
        }
      },
    });
  };

  const statusOptions = useMemo(
    () => [
      { label: t('全部状态'), value: '' },
      ...getTicketStatusOptions(t),
    ],
    [t],
  );

  const typeOptions = useMemo(
    () => [
      { label: t('全部类型'), value: '' },
      ...getTicketTypeOptions(t, { includeInvoice: true }),
    ],
    [t],
  );

  return (
    <div className='mt-[60px] px-2'>
      <CreateTicketModal
        visible={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={() => {
          setActivePage(1);
          loadTickets();
        }}
        t={t}
      />
      <CreateInvoiceTicketModal
        visible={showInvoiceModal}
        onClose={() => setShowInvoiceModal(false)}
        onSuccess={() => {
          setActivePage(1);
          loadTickets();
        }}
        t={t}
      />

      <TicketsPage
        title={t('工单中心')}
        description={t('在这里提交问题、查看处理进度，并与管理员继续沟通')}
        compactMode={compactMode}
        setCompactMode={setCompactMode}
        tickets={tickets}
        loading={loading}
        activePage={activePage}
        pageSize={pageSize}
        ticketCount={ticketCount}
        handlePageChange={setActivePage}
        handlePageSizeChange={(size) => {
          setPageSize(size);
          setActivePage(1);
        }}
        onOpenDetail={(ticket) => navigate(`/console/ticket/${ticket.id}`)}
        onCloseTicket={handleCloseTicket}
        t={t}
        actionsArea={
          <div className='flex flex-col md:flex-row md:items-center md:justify-between gap-3 w-full'>
            <Space wrap>
              <Button
                theme='solid'
                type='primary'
                icon={<IconPlusCircle />}
                onClick={() => setShowCreateModal(true)}
              >
                {t('新建工单')}
              </Button>
              <Button
                theme='light'
                type='primary'
                onClick={() => setShowInvoiceModal(true)}
              >
                {t('申请发票')}
              </Button>
            </Space>
            <Space wrap>
              <Select
                value={statusFilter}
                optionList={statusOptions}
                style={{ width: 160 }}
                onChange={(value) => {
                  setStatusFilter(value);
                  setActivePage(1);
                }}
              />
              <Select
                value={typeFilter}
                optionList={typeOptions}
                style={{ width: 160 }}
                onChange={(value) => {
                  setTypeFilter(value);
                  setActivePage(1);
                }}
              />
            </Space>
          </div>
        }
      />
    </div>
  );
};

export default Ticket;

