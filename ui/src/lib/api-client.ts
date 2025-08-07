import { Configuration } from '../generated/api/src/runtime';
import { ResponseError } from '../generated/api/src/runtime';
import { ErrorModelFromJSON } from '../generated/api/src/models/ErrorModel';
import { AuthenticationApi } from '../generated/api/src/apis/AuthenticationApi';
import { UsersApi } from '../generated/api/src/apis/UsersApi';
import { PseudonymsApi } from '../generated/api/src/apis/PseudonymsApi';
import { ContentApi } from '../generated/api/src/apis/ContentApi';
import { SubforumsApi } from '../generated/api/src/apis/SubforumsApi';
import { MessagesApi } from '../generated/api/src/apis/MessagesApi';
import { ModerationApi } from '../generated/api/src/apis/ModerationApi';
import { ReportsApi } from '../generated/api/src/apis/ReportsApi';
import { SearchApi } from '../generated/api/src/apis/SearchApi';
import { RulesApi } from '../generated/api/src/apis/RulesApi';
import { CorrelationApi } from '../generated/api/src/apis/CorrelationApi';
import { AdministrationApi } from '../generated/api/src/apis/AdministrationApi';

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8888';

// Helper to get a pre-configured API instance with authentication
export function getApi<T extends { new(config: Configuration): InstanceType<T> }>(ApiClass: T): InstanceType<T> {
  const config = new Configuration({
    basePath: API_BASE_URL,
    credentials: 'include', // Always send cookies (for JWT)
    headers: {
      'Content-Type': 'application/json',
    },
  });

  return new ApiClass(config);
}

// Helper to extract error messages from API responses
export async function extractApiErrorMessage(error: unknown): Promise<string> {
  if (error instanceof ResponseError) {
    try {
      const responseText = await error.response.text();
      const errorData = JSON.parse(responseText);
      
      // Try to parse as ErrorModel
      if (errorData.detail) {
        return errorData.detail;
      }
      
      // Fallback to title or generic message
      if (errorData.title) {
        return errorData.title;
      }
      
      return `Request failed with status ${error.response.status}`;
    } catch (parseError) {
      return `Request failed with status ${error.response.status}`;
    }
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return 'An unexpected error occurred';
}

// Export API instances directly
export const authApi = getApi(AuthenticationApi);
export const usersApi = getApi(UsersApi);
export const pseudonymsApi = getApi(PseudonymsApi);
export const contentApi = getApi(ContentApi);
export const subforumsApi = getApi(SubforumsApi);
export const messagesApi = getApi(MessagesApi);
export const moderationApi = getApi(ModerationApi);
export const reportsApi = getApi(ReportsApi);
export const searchApi = getApi(SearchApi);
export const rulesApi = getApi(RulesApi);
export const correlationApi = getApi(CorrelationApi);
export const administrationApi = getApi(AdministrationApi);