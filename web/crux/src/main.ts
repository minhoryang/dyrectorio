import { ServerCredentials } from '@grpc/grpc-js'
import { INestApplication, Logger, LogLevel } from '@nestjs/common'
import { ConfigService } from '@nestjs/config'
import { NestFactory } from '@nestjs/core'
import { MicroserviceOptions, Transport } from '@nestjs/microservices'
import { join } from 'path'
import AppModule from './app.module'
import ShutdownService from './application.shutdown.service'
import getServerCredentials from './shared/cert'

const HOUR_IN_MS: number = 60 * 60 * 1000

type GrpcOptions = {
  url: string
  credentials: ServerCredentials
}

type GrpcCertPrefix = 'api' | 'agent'

const loadGrpcOptions = (
  app: INestApplication,
  certPrefix: GrpcCertPrefix,
  portEnv: string,
  insecureEnv: string,
): GrpcOptions => {
  const port = portEnv ? Number(portEnv) : certPrefix === 'agent' ? 5000 : 5001

  const certPairs =
    insecureEnv === 'true'
      ? null
      : getServerCredentials(certPrefix, () => {
          app.get(ShutdownService).subscribeToShutdown(() => app.close())
        })

  return {
    credentials: certPairs ? ServerCredentials.createSsl(null, certPairs, false) : ServerCredentials.createInsecure(),
    url: `0.0.0.0:${port}`,
  }
}

const parseLogLevelFromEnv = (logLevelEnv: string, nodeEnv: string): LogLevel[] => {
  const VALID_LOG_LEVEL_VALUES = ['error', 'warn', 'log', 'debug', 'verbose']

  const index = VALID_LOG_LEVEL_VALUES.indexOf(logLevelEnv)
  if (index < 0) {
    return nodeEnv === 'production' ? ['error', 'warn'] : ['log', 'error', 'warn', 'debug']
  }

  return VALID_LOG_LEVEL_VALUES.slice(0, index) as LogLevel[]
}

const bootstrap = async () => {
  const app = await NestFactory.create(AppModule, {
    logger: parseLogLevelFromEnv(process.env.LOG_LEVEL, process.env.NODE_ENV),
  })

  app.enableShutdownHooks()

  const configService = app.get(ConfigService)

  const agentOptions = loadGrpcOptions(
    app,
    'agent',
    configService.get<string>('GRPC_AGENT_PORT'),
    configService.get<string>('GRPC_AGENT_INSECURE'),
  )
  const apiOptions = loadGrpcOptions(
    app,
    'api',
    configService.get<string>('GRPC_API_PORT'),
    configService.get<string>('GRPC_API_INSECURE'),
  )

  // agent
  app.connectMicroservice<MicroserviceOptions>({
    transport: Transport.GRPC,
    options: {
      package: ['agent'],
      protoPath: [
        join(__dirname, '../proto/crux.proto'),
        join(__dirname, '../proto/agent.proto'),
        join(__dirname, '../proto/common.proto'),
      ],
      keepalive: { keepaliveTimeoutMs: HOUR_IN_MS },
      ...agentOptions,
    },
  })

  // API
  app.connectMicroservice<MicroserviceOptions>(
    {
      transport: Transport.GRPC,
      options: {
        package: ['crux'],
        protoPath: [join(__dirname, '../proto/crux.proto'), join(__dirname, '../proto/common.proto')],
        keepalive: { keepaliveTimeoutMs: HOUR_IN_MS },
        ...apiOptions,
      },
    },
    { inheritAppConfig: true },
  )

  const logger: Logger = new Logger('NestBoostrap')
  await app.startAllMicroservices()

  logger.log(`gRPC agent services are running on: ${agentOptions.url}`)
  logger.log(`gRPC API services are running on: ${apiOptions.url}`)
}

bootstrap()
