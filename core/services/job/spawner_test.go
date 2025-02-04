package job_test

import (
	"context"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink/core/utils"
	"github.com/stretchr/testify/assert"

	"github.com/onsi/gomega"
	"gorm.io/gorm"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink/core/internal/cltest"
	"github.com/smartcontractkit/chainlink/core/internal/testutils/evmtest"
	"github.com/smartcontractkit/chainlink/core/internal/testutils/pgtest"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/services/eth"
	"github.com/smartcontractkit/chainlink/core/services/job"
	"github.com/smartcontractkit/chainlink/core/services/job/mocks"
	"github.com/smartcontractkit/chainlink/core/services/offchainreporting"
	"github.com/smartcontractkit/chainlink/core/services/pipeline"
	"github.com/smartcontractkit/chainlink/core/services/postgres"
)

type delegate struct {
	jobType                    job.Type
	services                   []job.Service
	jobID                      int32
	chContinueCreatingServices chan struct{}
	job.Delegate
}

func (d delegate) JobType() job.Type {
	return d.jobType
}

func (d delegate) ServicesForSpec(js job.Job) ([]job.Service, error) {
	if js.Type != d.jobType {
		return nil, nil
	}
	return d.services, nil
}

func clearDB(t *testing.T, db *gorm.DB) {
	err := db.Exec(`TRUNCATE jobs, pipeline_runs, pipeline_specs, pipeline_task_runs CASCADE`).Error
	require.NoError(t, err)
}

func TestSpawner_CreateJobDeleteJob(t *testing.T) {
	config := cltest.NewTestGeneralConfig(t)
	gdb := pgtest.NewGormDB(t)
	db := postgres.UnwrapGormDB(gdb)
	keyStore := cltest.NewKeyStore(t, db)
	ethKeyStore := keyStore.Eth()
	keyStore.OCR().Add(cltest.DefaultOCRKey)

	_, address := cltest.MustInsertRandomKey(t, ethKeyStore)
	_, bridge := cltest.MustCreateBridge(t, db, cltest.BridgeOpts{})
	_, bridge2 := cltest.MustCreateBridge(t, db, cltest.BridgeOpts{})

	ethClient, _, _ := cltest.NewEthMocksWithDefaultChain(t)
	ethClient.On("CallContext", mock.Anything, mock.Anything, "eth_getBlockByNumber", mock.Anything, false).
		Run(func(args mock.Arguments) {
			head := args.Get(1).(**eth.Head)
			*head = cltest.Head(10)
		}).
		Return(nil)
	cc := evmtest.NewChainSet(t, evmtest.TestChainOpts{DB: gdb, Client: ethClient, GeneralConfig: config})

	t.Run("should respect its dependents", func(t *testing.T) {
		lggr := logger.TestLogger(t)
		orm := job.NewTestORM(t, db, cc, pipeline.NewORM(db, lggr), keyStore)
		a := utils.NewDependentAwaiter()
		a.AddDependents(1)
		spawner := job.NewSpawner(orm, config, map[job.Type]job.Delegate{}, db, lggr, []utils.DependentAwaiter{a})
		// Starting the spawner should signal to the dependents
		result := make(chan bool)
		go func() {
			select {
			case <-a.AwaitDependents():
				result <- true
			case <-time.After(2 * time.Second):
				result <- false
			}
		}()
		spawner.Start()
		assert.True(t, <-result, "failed to signal to dependents")
	})

	t.Run("starts and stops job services when jobs are added and removed", func(t *testing.T) {
		jobA := cltest.MakeDirectRequestJobSpec(t)
		jobB := makeOCRJobSpec(t, address, bridge.Name.String(), bridge2.Name.String())

		lggr := logger.TestLogger(t)
		orm := job.NewTestORM(t, db, cc, pipeline.NewORM(db, lggr), keyStore)
		eventuallyA := cltest.NewAwaiter()
		serviceA1 := new(mocks.Service)
		serviceA2 := new(mocks.Service)
		serviceA1.On("Start").Return(nil).Once()
		serviceA2.On("Start").Return(nil).Once().Run(func(mock.Arguments) { eventuallyA.ItHappened() })
		dA := offchainreporting.NewDelegate(nil, orm, nil, nil, nil, monitoringEndpoint, cc, logger.TestLogger(t))
		delegateA := &delegate{jobA.Type, []job.Service{serviceA1, serviceA2}, 0, make(chan struct{}), dA}
		eventuallyB := cltest.NewAwaiter()
		serviceB1 := new(mocks.Service)
		serviceB2 := new(mocks.Service)
		serviceB1.On("Start").Return(nil).Once()
		serviceB2.On("Start").Return(nil).Once().Run(func(mock.Arguments) { eventuallyB.ItHappened() })

		dB := offchainreporting.NewDelegate(nil, orm, nil, nil, nil, monitoringEndpoint, cc, logger.TestLogger(t))
		delegateB := &delegate{jobB.Type, []job.Service{serviceB1, serviceB2}, 0, make(chan struct{}), dB}
		spawner := job.NewSpawner(orm, config, map[job.Type]job.Delegate{
			jobA.Type: delegateA,
			jobB.Type: delegateB,
		}, db, lggr, nil)
		spawner.Start()
		err := spawner.CreateJob(jobA)
		require.NoError(t, err)
		jobSpecIDA := jobA.ID
		delegateA.jobID = jobSpecIDA
		close(delegateA.chContinueCreatingServices)

		eventuallyA.AwaitOrFail(t, 20*time.Second)
		mock.AssertExpectationsForObjects(t, serviceA1, serviceA2)

		err = spawner.CreateJob(jobB)
		require.NoError(t, err)
		jobSpecIDB := jobB.ID
		delegateB.jobID = jobSpecIDB
		close(delegateB.chContinueCreatingServices)

		eventuallyB.AwaitOrFail(t, 20*time.Second)
		mock.AssertExpectationsForObjects(t, serviceB1, serviceB2)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		serviceA1.On("Close").Return(nil).Once()
		serviceA2.On("Close").Return(nil).Once()
		err = spawner.DeleteJob(ctx, jobSpecIDA)
		require.NoError(t, err)

		serviceB1.On("Close").Return(nil).Once()
		serviceB2.On("Close").Return(nil).Once()
		err = spawner.DeleteJob(ctx, jobSpecIDB)
		require.NoError(t, err)

		require.NoError(t, spawner.Close())
		serviceA1.AssertExpectations(t)
		serviceA2.AssertExpectations(t)
		serviceB1.AssertExpectations(t)
		serviceB2.AssertExpectations(t)
	})

	clearDB(t, gdb)

	t.Run("starts and stops job services from the DB when .Start()/.Stop() is called", func(t *testing.T) {
		jobA := makeOCRJobSpec(t, address, bridge.Name.String(), bridge2.Name.String())

		eventually := cltest.NewAwaiter()
		serviceA1 := new(mocks.Service)
		serviceA2 := new(mocks.Service)
		serviceA1.On("Start").Return(nil).Once()
		serviceA2.On("Start").Return(nil).Once().Run(func(mock.Arguments) { eventually.ItHappened() })

		lggr := logger.TestLogger(t)
		orm := job.NewTestORM(t, db, cc, pipeline.NewORM(db, lggr), keyStore)
		d := offchainreporting.NewDelegate(nil, orm, nil, nil, nil, monitoringEndpoint, cc, logger.TestLogger(t))
		delegateA := &delegate{jobA.Type, []job.Service{serviceA1, serviceA2}, 0, nil, d}
		spawner := job.NewSpawner(orm, config, map[job.Type]job.Delegate{
			jobA.Type: delegateA,
		}, db, lggr, nil)

		err := orm.CreateJob(jobA)
		require.NoError(t, err)
		delegateA.jobID = jobA.ID

		spawner.Start()

		eventually.AwaitOrFail(t)
		mock.AssertExpectationsForObjects(t, serviceA1, serviceA2)

		serviceA1.On("Close").Return(nil).Once()
		serviceA2.On("Close").Return(nil).Once()

		require.NoError(t, spawner.Close())

		mock.AssertExpectationsForObjects(t, serviceA1, serviceA2)
	})

	clearDB(t, gdb)

	t.Run("closes job services on 'DeleteJob()'", func(t *testing.T) {
		jobA := makeOCRJobSpec(t, address, bridge.Name.String(), bridge2.Name.String())

		eventuallyStart := cltest.NewAwaiter()
		serviceA1 := new(mocks.Service)
		serviceA2 := new(mocks.Service)
		serviceA1.On("Start").Return(nil).Once()
		serviceA2.On("Start").Return(nil).Once().Run(func(mock.Arguments) { eventuallyStart.ItHappened() })

		lggr := logger.TestLogger(t)
		orm := job.NewTestORM(t, db, cc, pipeline.NewORM(db, lggr), keyStore)
		d := offchainreporting.NewDelegate(nil, orm, nil, nil, nil, monitoringEndpoint, cc, logger.TestLogger(t))
		delegateA := &delegate{jobA.Type, []job.Service{serviceA1, serviceA2}, 0, nil, d}
		spawner := job.NewSpawner(orm, config, map[job.Type]job.Delegate{
			jobA.Type: delegateA,
		}, db, lggr, nil)

		err := orm.CreateJob(jobA)
		require.NoError(t, err)
		jobSpecIDA := jobA.ID
		delegateA.jobID = jobSpecIDA

		spawner.Start()
		defer spawner.Close()

		eventuallyStart.AwaitOrFail(t)

		// Wait for the claim lock to be taken
		gomega.NewGomegaWithT(t).Eventually(func() bool {
			jobs := spawner.ActiveJobs()
			_, exists := jobs[jobSpecIDA]
			return exists
		}, cltest.DBWaitTimeout, cltest.DBPollingInterval).Should(gomega.Equal(true))

		eventuallyClose := cltest.NewAwaiter()
		serviceA1.On("Close").Return(nil).Once()
		serviceA2.On("Close").Return(nil).Once().Run(func(mock.Arguments) { eventuallyClose.ItHappened() })

		err = spawner.DeleteJob(context.Background(), jobSpecIDA)
		require.NoError(t, err)

		eventuallyClose.AwaitOrFail(t)

		// Wait for the claim lock to be released
		gomega.NewGomegaWithT(t).Eventually(func() bool {
			jobs := spawner.ActiveJobs()
			_, exists := jobs[jobSpecIDA]
			return exists
		}, cltest.DBWaitTimeout, cltest.DBPollingInterval).Should(gomega.Equal(false))

		mock.AssertExpectationsForObjects(t, serviceA1, serviceA2)
	})
}
