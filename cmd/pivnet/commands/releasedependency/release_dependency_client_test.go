package releasedependency_test

import (
	"bytes"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pivnet "github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/releasedependency"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/releasedependency/releasedependencyfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler/errorhandlerfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

var _ = Describe("releasedependency commands", func() {
	var (
		fakePivnetClient *releasedependencyfakes.FakePivnetClient

		fakeErrorHandler *errorhandlerfakes.FakeErrorHandler

		outBuffer bytes.Buffer

		releasedependencies []pivnet.ReleaseDependency

		client *releasedependency.ReleaseDependencyClient
	)

	BeforeEach(func() {
		fakePivnetClient = &releasedependencyfakes.FakePivnetClient{}

		outBuffer = bytes.Buffer{}

		fakeErrorHandler = &errorhandlerfakes.FakeErrorHandler{}

		releasedependencies = []pivnet.ReleaseDependency{
			{
				Release: pivnet.DependentRelease{
					ID: 1234,
				},
			},
			{
				Release: pivnet.DependentRelease{
					ID: 2345,
				},
			},
		}

		fakePivnetClient.ReleaseDependenciesReturns(releasedependencies, nil)

		client = releasedependency.NewReleaseDependencyClient(
			fakePivnetClient,
			fakeErrorHandler,
			printer.PrintAsJSON,
			&outBuffer,
			printer.NewPrinter(&outBuffer),
		)
	})

	Describe("ReleaseDependencies", func() {
		var (
			productSlug    string
			releaseVersion string
		)

		BeforeEach(func() {
			productSlug = "some product slug"
			releaseVersion = "some release version"
		})

		It("lists all ReleaseDependencies", func() {
			err := client.List(productSlug, releaseVersion)
			Expect(err).NotTo(HaveOccurred())

			var returnedReleaseDependencies []pivnet.ReleaseDependency
			err = json.Unmarshal(outBuffer.Bytes(), &returnedReleaseDependencies)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedReleaseDependencies).To(Equal(releasedependencies))
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("releasedependencies error")
				fakePivnetClient.ReleaseDependenciesReturns(nil, expectedErr)
			})

			It("invokes the error handler", func() {
				err := client.List(productSlug, releaseVersion)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})

		Context("when there is an error getting release", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("releases error")
				fakePivnetClient.ReleaseForProductVersionReturns(pivnet.Release{}, expectedErr)
			})

			It("invokes the error handler", func() {
				err := client.List(productSlug, releaseVersion)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})
})
