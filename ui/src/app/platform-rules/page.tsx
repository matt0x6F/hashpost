import { PlatformRulesDisplay } from '@/components/PlatformRulesDisplay';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';

export default function PlatformRulesPage() {
  return (
    <div className="max-w-4xl mx-auto p-2 sm:p-4 space-y-6">
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold">Platform Rules & Guidelines</h1>
        <p className="text-muted-foreground">
          These rules apply to all content and behavior across HashPost. Please familiarize yourself with them to ensure a positive experience for everyone.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Community Guidelines</CardTitle>
          <CardDescription>
            Our platform is built on the principles of respectful discourse, privacy protection, and meaningful conversations. These guidelines help maintain the quality of our communities.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <h3 className="font-semibold text-foreground">Be Respectful</h3>
              <p className="text-sm text-muted-foreground">
                Treat others with kindness and respect, even when you disagree. Personal attacks, harassment, and hate speech have no place here.
              </p>
            </div>
            <div className="space-y-2">
              <h3 className="font-semibold text-foreground">Protect Privacy</h3>
              <p className="text-sm text-muted-foreground">
                Respect the privacy of others. Don&apos;t share personal information without consent, and be mindful of pseudonym boundaries.
              </p>
            </div>
            <div className="space-y-2">
              <h3 className="font-semibold text-foreground">Share Quality Content</h3>
              <p className="text-sm text-muted-foreground">
                Contribute meaningful, relevant content that adds value to discussions. Avoid spam, misinformation, and low-effort posts.
              </p>
            </div>
            <div className="space-y-2">
              <h3 className="font-semibold text-foreground">Follow the Law</h3>
              <p className="text-sm text-muted-foreground">
                Ensure your content complies with applicable laws. Illegal content, copyright violations, and harmful activities are prohibited.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Platform Rules</CardTitle>
          <CardDescription>
            These are the specific rules that govern content and behavior on HashPost. Violations may result in content removal, temporary restrictions, or account suspension.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <PlatformRulesDisplay showTitle={false} maxHeight="h-96" />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>How Rules Are Enforced</CardTitle>
          <CardDescription>
            Understanding how our moderation system works helps ensure fair and consistent enforcement.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-3">
            <div className="flex items-start gap-3">
              <div className="w-2 h-2 bg-blue-500 rounded-full mt-2 flex-shrink-0"></div>
              <div>
                <h4 className="font-medium">Community Moderation</h4>
                <p className="text-sm text-muted-foreground">
                  Subforum moderators handle most rule violations within their communities, following both platform rules and community-specific guidelines.
                </p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="w-2 h-2 bg-orange-500 rounded-full mt-2 flex-shrink-0"></div>
              <div>
                <h4 className="font-medium">Platform Escalation</h4>
                <p className="text-sm text-muted-foreground">
                  Serious violations or cross-community issues are escalated to platform moderators who have access to additional tools and broader context.
                </p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="w-2 h-2 bg-red-500 rounded-full mt-2 flex-shrink-0"></div>
              <div>
                <h4 className="font-medium">Appeals Process</h4>
                <p className="text-sm text-muted-foreground">
                  Users can appeal moderation decisions through our support system. We review appeals carefully and may overturn decisions when appropriate.
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Reporting Violations</CardTitle>
          <CardDescription>
            If you encounter content that violates our rules, please report it so we can take appropriate action.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-muted-foreground">
            When reporting content, please specify which rule was violated and provide context about why the content is problematic. This helps moderators make informed decisions.
          </p>
          <div className="bg-muted/50 p-3 rounded-lg">
            <p className="text-sm text-muted-foreground">
              <strong>Note:</strong> False reports waste moderator time and may result in consequences for the reporting user. Only report content that genuinely violates our rules.
            </p>
          </div>
        </CardContent>
      </Card>

      <div className="text-center text-sm text-muted-foreground">
        <p>
          These rules are updated periodically to reflect our evolving community standards. 
          For questions about specific rules or enforcement, please contact our support team.
        </p>
      </div>
    </div>
  );
}
