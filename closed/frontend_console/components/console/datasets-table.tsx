'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';

import { CopyButton } from '@/components/console/copy-button';
import { Button } from '@/components/ui/button';
import { Table, TableContainer, TableEmpty } from '@/components/ui/table';
import type { components } from '@/lib/gateway-openapi';

export type Dataset = components['schemas']['Dataset'];

export function DatasetsTable({ datasets }: { datasets: Dataset[] }) {
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState<'all' | 'with_description' | 'without_description'>('all');

  const filtered = useMemo(() => {
    if (!query.trim()) {
      return datasets.filter((dataset) => {
        if (filter === 'with_description') {
          return Boolean(dataset.description);
        }
        if (filter === 'without_description') {
          return !dataset.description;
        }
        return true;
      });
    }
    const q = query.trim().toLowerCase();
    return datasets.filter((dataset) =>
      [dataset.dataset_id, dataset.name, dataset.description ?? ''].some((value) => value.toLowerCase().includes(q)),
    );
  }, [datasets, query, filter]);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center gap-3">
        <input
          className="h-9 w-64 rounded-md border border-input bg-transparent px-3 text-sm"
          placeholder="Поиск по имени или идентификатору"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
        />
        <div className="console-pill">Сортировка: created_at ↓</div>
        <div className="flex flex-wrap gap-2">
          <Button variant={filter === 'all' ? 'default' : 'secondary'} size="sm" onClick={() => setFilter('all')}>
            Все
          </Button>
          <Button
            variant={filter === 'with_description' ? 'default' : 'secondary'}
            size="sm"
            onClick={() => setFilter('with_description')}
          >
            С описанием
          </Button>
          <Button
            variant={filter === 'without_description' ? 'default' : 'secondary'}
            size="sm"
            onClick={() => setFilter('without_description')}
          >
            Без описания
          </Button>
        </div>
      </div>
      <TableContainer>
        <Table>
          <thead>
            <tr>
              <th>Dataset ID</th>
              <th>Имя</th>
              <th>Описание</th>
              <th>Создан</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((dataset) => (
              <tr key={dataset.dataset_id}>
                <td className="font-mono text-xs">{dataset.dataset_id}</td>
                <td>
                  <Link href={`/console/datasets?dataset_id=${dataset.dataset_id}`} className="text-sm font-semibold text-primary">
                    {dataset.name}
                  </Link>
                </td>
                <td className="text-muted-foreground">{dataset.description ?? '—'}</td>
                <td className="text-xs text-muted-foreground">{dataset.created_at}</td>
                <td>
                  <CopyButton value={dataset.dataset_id} />
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
        {filtered.length === 0 ? <TableEmpty>Ничего не найдено.</TableEmpty> : null}
      </TableContainer>
    </div>
  );
}
