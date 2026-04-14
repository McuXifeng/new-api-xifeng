import React, { useMemo } from 'react';
import { Empty } from '@douyinfe/semi-ui';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import CardTable from '../../common/ui/CardTable';
import { getTicketsColumns } from './TicketsColumnDefs';

const TicketsTable = ({
  tickets = [],
  loading = false,
  compactMode = false,
  activePage,
  pageSize,
  ticketCount,
  handlePageChange,
  handlePageSizeChange,
  admin = false,
  onOpenDetail,
  onCloseTicket,
  t,
}) => {
  const columns = useMemo(
    () =>
      getTicketsColumns({
        t,
        admin,
        onOpenDetail,
        onCloseTicket,
      }),
    [t, admin, onOpenDetail, onCloseTicket],
  );

  return (
    <CardTable
      rowKey='id'
      columns={columns}
      dataSource={tickets}
      loading={loading}
      hidePagination
      scroll={undefined}
      pagination={{
        currentPage: activePage,
        pageSize,
        total: ticketCount,
        pageSizeOpts: [10, 20, 50, 100],
        showSizeChanger: true,
        onPageSizeChange: handlePageSizeChange,
        onPageChange: handlePageChange,
      }}
      onRow={(record) => ({
        onClick: () => onOpenDetail?.(record),
      })}
      empty={
        <Empty
          image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
          darkModeImage={
            <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
          }
          description={t('暂无工单')}
          style={{ padding: 30 }}
        />
      }
      className='overflow-hidden'
    />
  );
};

export default TicketsTable;

