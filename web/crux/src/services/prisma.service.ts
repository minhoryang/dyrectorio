import { INestApplication, Injectable, Logger, OnModuleInit } from '@nestjs/common'
import { Prisma, PrismaClient } from '@prisma/client'

@Injectable()
export default class PrismaService extends PrismaClient<Prisma.PrismaClientOptions, 'query'> implements OnModuleInit {
  private readonly logger = new Logger(PrismaService.name)

  constructor() {
    super({
      rejectOnNotFound: false,
      log: [
        { emit: 'stdout', level: 'warn' },
        { emit: 'stdout', level: 'error' },
      ],
    })
  }

  async onModuleInit() {
    await this.$connect()
  }

  async enableShutdownHooks(app: INestApplication) {
    this.$on('beforeExit', async () => {
      await app.close()
    })
  }

  async findLastMigration(): Promise<string> {
    const migrations: { migration_name: string }[] = await this
      .$queryRaw`SELECT migration_name from _prisma_migrations WHERE rolled_back_at IS NULL ORDER BY finished_at DESC LIMIT 1`

    return migrations && migrations.length > 0 ? migrations[0].migration_name : null
  }
}
