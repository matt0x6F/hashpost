'use client';

import { Card, CardContent } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { ArrowLeft, Lock, UserX } from 'lucide-react';
import { useRouter } from 'next/navigation';

interface UnauthorizedPageProps {
  title?: string;
  message?: string;
  showBackButton?: boolean;
  backButtonText?: string;
  icon?: 'lock' | 'user-x';
}

export function UnauthorizedPage({
  title = 'Access Denied',
  message = 'You do not have permission to access this page.',
  showBackButton = true,
  backButtonText = 'Go Back',
  icon = 'lock',
}: UnauthorizedPageProps) {
  const router = useRouter();

  const IconComponent = icon === 'user-x' ? UserX : Lock;

  return (
    <div className="container mx-auto py-6 max-w-2xl">
      <Card>
        <CardContent className="p-8 text-center">
          <div className="flex flex-col items-center gap-4">
            <div className="rounded-full bg-muted p-4">
              <IconComponent className="h-8 w-8 text-muted-foreground" />
            </div>
            
            <div className="space-y-2">
              <h1 className="text-2xl font-bold">{title}</h1>
              <p className="text-muted-foreground">{message}</p>
            </div>

            {showBackButton && (
              <Button 
                onClick={() => router.back()} 
                variant="outline"
                className="mt-4"
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                {backButtonText}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
