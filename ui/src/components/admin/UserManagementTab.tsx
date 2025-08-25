"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Badge } from "@/components/shadcn/badge";
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from "@/components/shadcn/table";
import { 
  Search, 
  Filter, 
  MoreHorizontal, 
  UserPlus, 
  Shield,
  Eye,
  Edit
} from "lucide-react";
import { toast } from "sonner";
import { getApi } from "@/lib/api-client";
import { SearchApi } from "@/generated/api/src/apis/SearchApi";
import { SearchPseudonym } from "@/generated/api/src/models/SearchPseudonym";


export function UserManagementTab() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [pseudonyms, setPseudonyms] = useState<SearchPseudonym[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  // Restore search query from URL params
  useEffect(() => {
    const queryFromUrl = searchParams.get('query');
    if (queryFromUrl && queryFromUrl !== searchQuery) {
      setSearchQuery(queryFromUrl);
      // Auto-search if query is in URL
      if (queryFromUrl.trim()) {
        searchPseudonymsWithQuery(queryFromUrl);
      }
    }
  }, [searchParams]);

  const searchPseudonymsWithQuery = async (query: string) => {
    if (!query.trim()) return;
    
    setIsLoading(true);
    try {
      const api = getApi(SearchApi);
      const response = await api.searchPseudonyms(query);
      
      // Debug: Log the actual response to see what fields we're getting
      		// Debug: search API response and pseudonym data logged
      
      if (response.pseudonyms) {
        setPseudonyms(response.pseudonyms);
        toast.success(`Found ${response.pseudonyms.length} pseudonyms`);
      } else {
        setPseudonyms([]);
        toast.info("No pseudonyms found");
      }
    } catch (error: unknown) {
      console.error("Failed to search pseudonyms:", error);
      
      // Handle specific error cases
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'status' in error.response) {
        const status = (error.response as { status: number }).status;
        if (status === 401) {
          toast.error("Authentication required. Please log in again.");
        } else if (status === 403) {
          toast.error("Insufficient permissions for pseudonym search");
        } else if (status === 500) {
          toast.error("Server error. Please try again later.");
        } else {
          toast.error("Failed to search pseudonyms. Please try again.");
        }
      } else {
        toast.error("Failed to search pseudonyms. Please try again.");
      }
      
      setPseudonyms([]);
    } finally {
      setIsLoading(false);
    }
  };

  const searchPseudonyms = async () => {
    await searchPseudonymsWithQuery(searchQuery);
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Update URL with search query
    const newParams = new URLSearchParams(searchParams);
    if (searchQuery.trim()) {
      newParams.set('query', searchQuery);
    } else {
      newParams.delete('query');
    }
    newParams.delete('page'); // Reset to first page on new search
    
    const newUrl = `/admin?tab=users${newParams.toString() ? `&${newParams.toString()}` : ''}`;
    router.replace(newUrl, { scroll: false });
    
    searchPseudonyms();
  };

  const handleViewPseudonym = (pseudonym: SearchPseudonym) => {
    // Debug: Log the pseudonym data we're about to store
    		// Debug: handleViewPseudonym called with pseudonym data
    
    // Store pseudonym data in sessionStorage to avoid URL parameter truncation
    const pseudonymDataKey = `pseudonym_detail_${pseudonym.pseudonymId}`;
    const pseudonymDataString = JSON.stringify(pseudonym);
    		// Debug: storing pseudonym data in sessionStorage
    
    sessionStorage.setItem(pseudonymDataKey, pseudonymDataString);
    
    // Verify it was stored
    sessionStorage.getItem(pseudonymDataKey);
    		// Debug: retrieved data from sessionStorage
    
    // Navigate to pseudonym detail page with search context preserved
    const searchParams = new URLSearchParams();
    searchParams.set('query', searchQuery);
    searchParams.set('page', '1'); // You could add pagination later
    
    const pseudonymDetailUrl = `/admin/pseudonyms/${pseudonym.pseudonymId}?${searchParams.toString()}`;
    		// Debug: navigating to pseudonym detail page
    router.push(pseudonymDetailUrl);
  };

  const getStatusBadge = (pseudonym: SearchPseudonym) => {
    if (pseudonym.isActive) {
      return <Badge variant="secondary">Active</Badge>;
    } else {
      return <Badge variant="destructive">Inactive</Badge>;
    }
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleDateString();
  };

  return (
    <div className="space-y-6">
      {/* Search Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Search className="h-5 w-5" />
            Search Pseudonyms
          </CardTitle>
          <CardDescription>
            Search for pseudonyms by display name, slug, or bio
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSearch} className="flex gap-2">
            <Input
              placeholder="Search by display name, slug, or bio..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="flex-1"
            />
            <Button type="submit" disabled={isLoading || !searchQuery.trim()}>
              {isLoading ? "Searching..." : "Search"}
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Pseudonym List */}
      {pseudonyms.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span>Search Results ({pseudonyms.length})</span>
              <Button variant="outline" size="sm">
                <Filter className="h-4 w-4 mr-2" />
                Filter
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Pseudonym</TableHead>
                  <TableHead>Karma Score</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {pseudonyms.map((pseudonym) => (
                  <TableRow key={pseudonym.pseudonymId}>
                    <TableCell>
                      <div>
                        <div className="font-medium">{pseudonym.displayName}</div>
                        <div className="text-sm text-muted-foreground">ID: {pseudonym.pseudonymId}</div>
                        {pseudonym.slug && (
                          <div className="text-sm text-muted-foreground">Slug: {pseudonym.slug}</div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{pseudonym.karmaScore}</Badge>
                    </TableCell>
                    <TableCell>
                      {getStatusBadge(pseudonym)}
                    </TableCell>
                    <TableCell>
                      {formatDate(pseudonym.createdAt)}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button 
                          variant="ghost" 
                          size="sm"
                          onClick={() => handleViewPseudonym(pseudonym)}
                          title="View pseudonym details"
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="sm">
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="sm">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
          <CardDescription>
            Common administrative tasks
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Button variant="outline" className="h-20 flex-col gap-2">
              <UserPlus className="h-6 w-6" />
              <span>Create Admin User</span>
            </Button>
            
            <Button variant="outline" className="h-20 flex-col gap-2">
              <Shield className="h-6 w-6" />
              <span>Manage Roles</span>
            </Button>
            
            <Button variant="outline" className="h-20 flex-col gap-2">
              <Eye className="h-6 w-6" />
              <span>View Reports</span>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
