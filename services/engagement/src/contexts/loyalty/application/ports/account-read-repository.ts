import type {PagedQueryPort} from '../../../../foundation/pagination.js';
import type {AccountView} from '../read-models/account.js';

/** Read accounts and their stable revisions through the loyalty boundary. */
export interface AccountReadRepository extends PagedQueryPort<AccountView> {}
