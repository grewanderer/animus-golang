import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { PageHeader, PageShell } from '@/components/ui/page-shell';

export default function ProjectsPage() {
  return (
    <PageShell>
      <PageHeader
        title="Проекты"
        description="Контекст доступа, управление ролями и архивирование. Все операции выполняются через Gateway API."
      />
      <Card>
        <CardHeader>
          <CardTitle>Список проектов</CardTitle>
          <CardDescription>Будет загружен из Gateway. Управление доступом сохраняет аудит и RBAC‑инварианты.</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Инициализация рабочего контура выполнена. Подключение к данным будет добавлено в следующих коммитах.
          </p>
        </CardContent>
      </Card>
    </PageShell>
  );
}
