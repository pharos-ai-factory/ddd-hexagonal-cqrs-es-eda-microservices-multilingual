import {Pool} from 'pg';
import schema from './generated/persistence.json' with {type: 'json'};
import contextSchemas from './generated/context-persistence.json' with {type: 'json'};

/** Owns one context pool and verifies its database identity, grants and migration ledger. */
export class PostgresContextDatabase {
  readonly pool: Pool;
  constructor(readonly owner: string, url: string) {
    this.pool = new Pool({connectionString: url, max: 6, connectionTimeoutMillis: 3000});
    // The pool removes failed idle clients; queries can acquire a replacement.
    this.pool.on('error', () => console.warn('Idle PostgreSQL connection lost for', this.owner));
  }
  async verify() {
    const {rows: [row]} = await this.pool.query(`SELECT current_database() AS database,
      rolsuper OR rolcreatedb OR rolcreaterole OR has_database_privilege(current_user,current_database(),'CREATE')
      OR has_schema_privilege(current_user,'cafe','CREATE') AS privileged
      FROM pg_roles WHERE rolname=current_user`);
    if (row.database !== 'cafe_'+this.owner || row.privileged) throw new Error('Runtime owner or privilege mismatch');
    for (const [version, table] of [[1, 'schema_version'], [2, 'schema_migrations']] as const) {
      const {rows: [migration]} = await this.pool.query(`SELECT checksum FROM cafe.${table} WHERE version=$1`, [version]);
      if (migration?.checksum !== schema[String(version) as keyof typeof schema]) throw new Error('Database migration checksum mismatch');
    }
    const {rows: [identity]} = await this.pool.query('SELECT owner FROM cafe.context_identity WHERE singleton');
    const expected = (contextSchemas as Record<string, Record<string, string>>)[this.owner];
    if (!expected || identity?.owner !== this.owner) throw new Error('Context schema identity mismatch');
    const {rows: migrations} = await this.pool.query<{version: number; checksum: string}>(
      'SELECT version,checksum FROM cafe.context_migrations');
    if (migrations.length !== Object.keys(expected).length || migrations.some(item => expected[String(item.version)] !== item.checksum)) {
      throw new Error('Context migration checksum mismatch');
    }
    await this.pool.query('SELECT id FROM cafe.realtime_publications LIMIT 0');
  }
}
