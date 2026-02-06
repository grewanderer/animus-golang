'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useOperations } from '@/lib/operations';
import { formatDateTime } from '@/lib/format';

const statusLabel: Record<string, { label: string; variant: 'info' | 'success' | 'warning' | 'neutral' }> = {
  pending: { label: 'В работе', variant: 'info' },
  succeeded: { label: 'Завершено', variant: 'success' },
  failed: { label: 'Ошибка', variant: 'warning' },
};

export function OperationsPanel() {
  const { operations, clearCompleted, retryOperation } = useOperations();

  if (operations.length === 0) {
    return null;
  }

  return (
    <Card className="border border-border/70">
      <CardHeader className="flex flex-row items-center justify-between gap-4">
        <div>
          <CardTitle>Фоновые операции</CardTitle>
          <CardDescription>
            Панель отслеживания запусков, экспортов и сеансов. Статусы фиксируются детерминированно.
          </CardDescription>
        </div>
        <Button variant="secondary" size="sm" onClick={clearCompleted}>
          Очистить завершённые
        </Button>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-3">
          {operations.map((operation) => {
            const meta = statusLabel[operation.status] ?? statusLabel.pending;
            return (
              <div key={operation.id} className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-border/60 px-3 py-3">
                <div className="min-w-[240px]">
                  <div className="text-sm font-semibold">{operation.label}</div>
                  <div className="text-xs text-muted-foreground">
                    Создано: {formatDateTime(operation.createdAt)} · Обновлено: {formatDateTime(operation.updatedAt ?? operation.createdAt)}
                  </div>
                  {operation.details ? <div className="text-xs text-muted-foreground">Контекст: {operation.details}</div> : null}
                  {operation.poll ? (
                    <div className="text-xs text-muted-foreground">Авто‑проверка: каждые 8с</div>
                  ) : null}
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={meta.variant}>{meta.label}</Badge>
                  {operation.status === 'failed' && operation.retry ? (
                    <Button variant="secondary" size="sm" onClick={() => retryOperation(operation)}>
                      Повторить
                    </Button>
                  ) : null}
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
