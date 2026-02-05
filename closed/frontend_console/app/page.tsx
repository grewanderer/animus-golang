import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export default function LandingPage() {
  return (
    <main className="mx-auto flex min-h-screen max-w-3xl items-center justify-center px-6 py-16">
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Консоль управления Animus Datalab</CardTitle>
          <CardDescription>
            Базовая оболочка и токены интерфейса готовы. Навигация и рабочие контуры будут доступны в следующих
            разделах.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Используйте прямой переход в рабочие разделы после включения маршрутизации. Контрольная плоскость остаётся
            источником истины, клиент взаимодействует только через Gateway API.
          </p>
        </CardContent>
      </Card>
    </main>
  );
}
