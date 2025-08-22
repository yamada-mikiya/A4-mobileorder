import { Button } from "@/components/ui/button";
import {
  Card,
  CardHeader,
  CardFooter,
  CardTitle,
  CardAction,
  CardDescription,
  CardContent,
} from "@/components/ui/card";

export default function Home() {
  return (
    <main className="min-h-screen p-8">
      <h1 className="text-4xl font-bold text-secondary">
        メニュー
      </h1>
      <div className="mt-8 space-y-4">
        <Card className="max-w-sm bg-white">
          <CardHeader>
            <CardTitle>たこ焼き</CardTitle>
            <CardDescription>おいしいです</CardDescription>
          </CardHeader>
          <CardContent>
            This is some content inside the card.
          </CardContent>
          <CardFooter>
            <CardAction>
              <Button variant="outline">Action</Button>
            </CardAction>
          </CardFooter>
        </Card>
                <Card className="max-w-sm bg-white">
          <CardHeader>
            <CardTitle>たこ焼き</CardTitle>
            <CardDescription>おいしいです</CardDescription>
          </CardHeader>
          <CardContent>
            This is some content inside the card.
          </CardContent>
          <CardFooter>
            <CardAction>
              <Button variant="outline">Action</Button>
            </CardAction>
          </CardFooter>
        </Card>
      </div>
    </main>
  );
}